package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/levinOo/geo-incedent-service/config"
	"github.com/levinOo/geo-incedent-service/internal/entity"
	"github.com/levinOo/geo-incedent-service/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIncidentService_Create(t *testing.T) {
	type args struct {
		ctx context.Context
		req *entity.CreateIncidentRequest
	}

	validReq := &entity.CreateIncidentRequest{
		Name:        "Fire",
		Description: "Big fire",
		Area: entity.GeoJsonPolygon{
			Type:        "Polygon",
			Coordinates: [][][]float64{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}},
		},
	}

	tests := []struct {
		name    string
		args    args
		mock    func(r *mocks.IncidentRepo)
		want    *entity.IncidentResponse
		wantErr bool
	}{
		{
			name: "Success",
			args: args{
				ctx: context.Background(),
				req: validReq,
			},
			mock: func(r *mocks.IncidentRepo) {
				r.On("Create", mock.Anything, mock.MatchedBy(func(i *entity.Incident) bool {
					return i.Name == validReq.Name && i.Description == validReq.Description
				})).Return(nil)
			},
			want: &entity.IncidentResponse{
				Status: "успешно создан",
			},
			wantErr: false,
		},
		{
			name: "Invalid Polygon",
			args: args{
				ctx: context.Background(),
				req: &entity.CreateIncidentRequest{
					Name:        "Fire",
					Description: "Big fire",
					Area: entity.GeoJsonPolygon{
						Type:        "Polygon",
						Coordinates: [][][]float64{{{0, 0}, {0, 10}}},
					},
				},
			},
			mock:    func(r *mocks.IncidentRepo) {},
			wantErr: true,
		},
		{
			name: "Repo Error",
			args: args{
				ctx: context.Background(),
				req: validReq,
			},
			mock: func(r *mocks.IncidentRepo) {
				r.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewIncidentRepo(t)
			tt.mock(repo)

			s := NewIncidentService(repo, &config.Config{})
			got, err := s.Create(tt.args.ctx, tt.args.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestIncidentService_FindByID(t *testing.T) {
	id := uuid.New()
	testIncident := &entity.Incident{
		ID:          id,
		Name:        "Test",
		Description: "Desc",
		Area:        entity.GeoJsonPolygon{},
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name    string
		id      string
		mock    func(r *mocks.IncidentRepo)
		want    *entity.GetIncidentResponse
		wantErr bool
	}{
		{
			name: "Success",
			id:   id.String(),
			mock: func(r *mocks.IncidentRepo) {
				r.On("FindByID", mock.Anything, id).Return(testIncident, nil)
			},
			want: &entity.GetIncidentResponse{
				ID:          id.String(),
				Name:        testIncident.Name,
				Description: testIncident.Description,
				Area:        testIncident.Area,
				IsActive:    testIncident.IsActive,
				CreatedAt:   testIncident.CreatedAt,
				UpdatedAt:   testIncident.UpdatedAt,
			},
			wantErr: false,
		},
		{
			name:    "Invalid UUID",
			id:      "invalid",
			mock:    func(r *mocks.IncidentRepo) {},
			wantErr: true,
		},
		{
			name: "Not Found",
			id:   id.String(),
			mock: func(r *mocks.IncidentRepo) {
				r.On("FindByID", mock.Anything, id).Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewIncidentRepo(t)
			tt.mock(repo)

			s := NewIncidentService(repo, &config.Config{})
			got, err := s.FindByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestIncidentService_Update(t *testing.T) {
	id := uuid.New()
	existing := &entity.Incident{
		ID:          id,
		Name:        "Old Name",
		Description: "Old Desc",
		IsActive:    true,
	}
	newName := "New Name"

	tests := []struct {
		name    string
		id      string
		req     *entity.UpdateIncidentRequest
		mock    func(r *mocks.IncidentRepo)
		want    *entity.IncidentResponse
		wantErr bool
	}{
		{
			name: "Success",
			id:   id.String(),
			req: &entity.UpdateIncidentRequest{
				Name: &newName,
			},
			mock: func(r *mocks.IncidentRepo) {
				r.On("FindByID", mock.Anything, id).Return(existing, nil)
				r.On("Update", mock.Anything, mock.MatchedBy(func(i *entity.Incident) bool {
					return i.Name == newName && i.Description == "Old Desc"
				})).Return(nil)
			},
			want:    &entity.IncidentResponse{Status: "успешно обновлен"},
			wantErr: false,
		},
		{
			name:    "No fields to update",
			id:      id.String(),
			req:     &entity.UpdateIncidentRequest{},
			mock:    func(r *mocks.IncidentRepo) {},
			wantErr: true,
		},
		{
			name: "Not Found",
			id:   id.String(),
			req:  &entity.UpdateIncidentRequest{Name: &newName},
			mock: func(r *mocks.IncidentRepo) {
				r.On("FindByID", mock.Anything, id).Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewIncidentRepo(t)
			tt.mock(repo)

			s := NewIncidentService(repo, &config.Config{})
			got, err := s.Update(context.Background(), tt.req, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestIncidentService_GetStats(t *testing.T) {
	tests := []struct {
		name     string
		mock     func(r *mocks.IncidentRepo)
		settings int
		want     *entity.StatsResponse
		wantErr  bool
	}{
		{
			name:     "Success",
			settings: 60,
			mock: func(r *mocks.IncidentRepo) {
				r.On("GetStats", mock.Anything, 60).Return([]*entity.IncidentStats{
					{Name: "Zone 1", UserCount: 10},
				}, nil)
			},
			want: &entity.StatsResponse{
				Stats: []*entity.IncidentStats{
					{Name: "Zone 1", UserCount: 10},
				},
				WindowMinutes: 60,
			},
			wantErr: false,
		},
		{
			name:     "Repo Error",
			settings: 30,
			mock: func(r *mocks.IncidentRepo) {
				r.On("GetStats", mock.Anything, 30).Return(nil, errors.New("err"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewIncidentRepo(t)
			tt.mock(repo)

			cfg := &config.Config{
				HTTPServer: config.HTTPServerConfig{StatsWindowMinutes: tt.settings},
			}
			s := NewIncidentService(repo, cfg)
			got, err := s.GetStats(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
