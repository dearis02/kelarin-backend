package cron

import (
	"context"
	"kelarin/internal/config"
	"sync"
	"time"

	robfigCron "github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

type Cron interface {
	RegisterJob(ctx context.Context, job config.Job, taskFn CronjobFunc) error
	Start()
	Stop()
}

type CronjobFunc func(ctx context.Context) error

func (f CronjobFunc) Run(ctx context.Context) error {
	return f(ctx)
}

type cronImpl struct {
	mu            sync.Mutex
	done          chan struct{}
	logger        config.CronLogger
	cronWithSkip  *robfigCron.Cron
	cronWithDelay *robfigCron.Cron
}

func New() Cron {
	c := &cronImpl{
		done: make(chan struct{}),
	}

	loc := time.FixedZone("UTC+8", 8*60*60)
	c.logger = config.NewCronLogger()

	c.cronWithSkip = robfigCron.New(
		robfigCron.WithLocation(loc),
		robfigCron.WithLogger(c.logger),
		robfigCron.WithChain(
			robfigCron.Recover(c.logger),
			robfigCron.SkipIfStillRunning(c.logger),
		),
	)

	c.cronWithDelay = robfigCron.New(
		robfigCron.WithLocation(loc),
		robfigCron.WithLogger(c.logger),
		robfigCron.WithChain(
			robfigCron.Recover(c.logger),
			robfigCron.DelayIfStillRunning(c.logger),
		),
	)

	return c
}

func (c *cronImpl) RegisterJob(ctx context.Context, job config.Job, jobFn CronjobFunc) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ctx = log.With().Str("job", job.Name).Logger().WithContext(ctx)
	log := log.Ctx(ctx)

	var entryID robfigCron.EntryID
	var err error

	switch job.ConcurrencyPolicy {
	case config.CronjobConcurrencyPolicySkip:
		entryID, err = c.cronWithSkip.AddFunc(job.Schedule, func() {
			err := jobFn.Run(ctx)
			if err != nil {
				log.Error().Err(err).Send()
			} else {
				log.Info().Msg("job executed successfully")
			}
		})
		if err != nil {
			return err
		}

	case config.CronjobConcurrencyPolicyWait:
		entryID, err = c.cronWithDelay.AddFunc(job.Schedule, func() {
			err := jobFn.Run(ctx)
			if err != nil {
				log.Error().Err(err).Send()
			} else {
				log.Info().Msg("job executed successfully")
			}
		})
		if err != nil {
			return err
		}

	}

	c.logger.AddEntry(entryID, job.Name)

	return nil
}

func (c *cronImpl) Start() {
	if c.cronWithSkip != nil {
		c.cronWithSkip.Start()
	}
	if c.cronWithDelay != nil {
		c.cronWithDelay.Start()
	}

	<-c.done
	close(c.done)
}

func (c *cronImpl) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	wg := sync.WaitGroup{}
	if c.cronWithSkip != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ctx := c.cronWithSkip.Stop()
			<-ctx.Done()
		}()
	}

	if c.cronWithDelay != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ctx := c.cronWithDelay.Stop()
			<-ctx.Done()
		}()
	}

	wg.Wait()
	c.done <- struct{}{}
}
