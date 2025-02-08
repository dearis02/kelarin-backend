package service

import (
	"context"
	"fmt"
	"kelarin/internal/repository"
	"kelarin/internal/types"
	pkg "kelarin/pkg/utils"
	"net/http"
	"path/filepath"

	"github.com/go-errors/errors"
	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/volatiletech/null/v9"
)

type ServiceProvider interface {
	Register(ctx context.Context, req types.ServiceProviderCreateReq) error
}

type serviceProviderImpl struct {
	db                      *sqlx.DB
	serviceProviderRepo     repository.ServiceProvider
	provinceRepo            repository.Province
	cityRepo                repository.City
	serviceProviderAreaRepo repository.ServiceProviderArea
	userRepo                repository.User
	pendingRegistrationRepo repository.PendingRegistration
	fileSvc                 File
	geocodingSvc            Geocoding
}

func NewServiceProvider(db *sqlx.DB, serviceProviderRepo repository.ServiceProvider, userRepo repository.User, provinceRepo repository.Province, cityRepo repository.City, ServiceProviderAreaRepo repository.ServiceProviderArea, pendingRegistrationRepo repository.PendingRegistration, fileSvc File, geocodingSvc Geocoding) ServiceProvider {
	return &serviceProviderImpl{
		db:                      db,
		serviceProviderRepo:     serviceProviderRepo,
		provinceRepo:            provinceRepo,
		cityRepo:                cityRepo,
		serviceProviderAreaRepo: ServiceProviderAreaRepo,
		userRepo:                userRepo,
		pendingRegistrationRepo: pendingRegistrationRepo,
		fileSvc:                 fileSvc,
		geocodingSvc:            geocodingSvc,
	}
}

func (s *serviceProviderImpl) Register(ctx context.Context, req types.ServiceProviderCreateReq) error {
	if err := req.Validate(); err != nil {
		return err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return errors.New(err)
	}

	user, err := s.userRepo.FindByID(ctx, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return errors.New("user not found")
	} else if err != nil {
		return err
	}

	svcProvider, err := s.serviceProviderRepo.FindByUserID(ctx, user.ID)
	if errors.Is(err, types.ErrNoData) {
		// do nothing
	} else if err != nil {
		return err
	}

	if svcProvider.ID != uuid.Nil {
		return errors.New(types.AppErr{
			Code:    http.StatusForbidden,
			Message: "service provider already registered",
		})
	}

	serviceProvider := types.ServiceProvider{
		ID:                id,
		UserID:            user.ID,
		Name:              req.Name,
		Description:       req.Description,
		HasPhysicalOffice: req.HasPhysicalOffice,
		Address:           req.Address,
		MobilePhoneNumber: req.MobilePhoneNumber,
		Telephone:         req.Telephone,
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.New(err)
	}

	defer tx.Rollback()

	if req.HasPhysicalOffice {
		lat := req.OfficeCoordinates[0].InexactFloat64()
		long := req.OfficeCoordinates[1].InexactFloat64()

		coordinates := s2.LatLngFromDegrees(lat, long)
		serviceProvider.OfficeCoordinates = null.StringFrom(fmt.Sprintf("POINT(%f %f)", lat, long))

		geocodingResChan := make(chan types.GeocodingReverseRes)
		errChan := make(chan error)
		go func() {
			res, err := s.geocodingSvc.Reverse(ctx, types.GeocodingReverseReq{LatLong: coordinates})
			errChan <- err
			close(errChan)

			geocodingResChan <- res
			close(geocodingResChan)
		}()

		if err := <-errChan; err != nil {
			return err
		}

		geocodingRes := <-geocodingResChan
		serviceProvider.Address = geocodingRes.Results[0].Formatted

		province, err := s.provinceRepo.FindByName(ctx, geocodingRes.Results[0].Components.State)
		if errors.Is(err, types.ErrNoData) {
		} else if err != nil {
			return errors.New(err)
		}

		var geoCodingResCity string
		if geocodingRes.Results[0].Components.City != "" {
			geoCodingResCity = geocodingRes.Results[0].Components.City
		} else if geocodingRes.Results[0].Components.County != "" {
			geoCodingResCity = geocodingRes.Results[0].Components.County
		}

		city, err := s.cityRepo.FindByProvinceIDAndName(ctx, province.ID, geoCodingResCity)
		if errors.Is(err, types.ErrNoData) {
		} else if err != nil {
			return errors.New(err)
		}

		if province.ID != 0 && city.ID != 0 {
			serviceProviderArea := types.ServiceProviderArea{
				ServiceProviderID: serviceProvider.ID,
				ProvinceID:        province.ID,
				CityID:            city.ID,
			}

			if err := s.serviceProviderAreaRepo.Create(ctx, tx, serviceProviderArea); err != nil {
				return err
			}
		}
	} else {
		province, err := s.provinceRepo.FindByID(ctx, req.ProvinceID.Int64)
		if errors.Is(err, types.ErrNoData) {
			log.Error().Int64("province_id", req.ProvinceID.Int64).Err(errors.New("province not found")).Send()
			return errors.New(types.AppErr{Code: http.StatusNotFound, Message: "province not found"})
		} else if err != nil {
			return err
		}

		city, err := s.cityRepo.FindByIDandProvinceID(ctx, req.CityID.Int64, province.ID)
		if errors.Is(err, types.ErrNoData) {
			log.Error().Int64("city_id", req.CityID.Int64).Int64("province_id", province.ID).Err(errors.New("city not found")).Send()
			return errors.New(types.AppErr{Code: http.StatusNotFound, Message: "city not found"})
		}

		serviceProviderArea := types.ServiceProviderArea{
			ServiceProviderID: serviceProvider.ID,
			ProvinceID:        province.ID,
			CityID:            city.ID,
		}

		if err := s.serviceProviderAreaRepo.Create(ctx, tx, serviceProviderArea); err != nil {
			return err
		}
	}

	tempFile, err := s.fileSvc.GetTemp(ctx, req.Logo)
	if err != nil {
		return err
	}

	if !pkg.IsFileExist(filepath.Join(types.TempFileDir, tempFile.Name)) {
		return errors.New(types.AppErr{
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("file %s not found", tempFile.Name),
		})
	}

	logo, err := s.fileSvc.BulkUploadToS3(ctx, []types.TempFile{{Name: tempFile.Name}}, types.ServiceProviderLogoDir)
	if err != nil {
		return err
	}

	if err = s.fileSvc.DeleteTemp(ctx, req.Logo); err != nil {
		return err
	}

	serviceProvider.LogoImage = logo[0]

	if err = s.serviceProviderRepo.Create(ctx, tx, serviceProvider); err != nil {
		return err
	}

	key := types.GetPendingRegistrationKey(req.AuthUser.ID.String())
	if err = s.pendingRegistrationRepo.Delete(ctx, key); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return errors.New(err)
	}

	return nil
}
