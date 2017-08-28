package soapboxd

import (
	"time"

	"github.com/adhocteam/soapbox/models"
	"github.com/adhocteam/soapbox/proto"
	pb "github.com/adhocteam/soapbox/proto"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

func (s *server) AddActivity(ctx context.Context, activity *pb.Activity) (*pb.Empty, error) {
	translations := map[pb.ActivityType]models.ActivityType{
		pb.ActivityType_APPLICATION_CREATED: models.ActivityTypeApplicationCreated,
		pb.ActivityType_DEPLOYMENT_STARTED:  models.ActivityTypeDeploymentStarted,
		pb.ActivityType_DEPLOYMENT_SUCCESS:  models.ActivityTypeDeploymentSuccess,
		pb.ActivityType_DEPLOYMENT_FAILURE:  models.ActivityTypeDeploymentFailure,
		pb.ActivityType_ENVIRONMENT_CREATED: models.ActivityTypeEnvironmentCreated,
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "beginning transaction")
	}

	insertQuery := `
	INSERT INTO activities (user_id, activity, application_id, deployment_id, environment_id, configuration_id) 
	VALUES ($1, $2, $3, $4, $5, $6)
	`
	var createdAt time.Time
	var activityType = translations[activity.Type]
	err = tx.QueryRow(insertQuery,
		activity.UserId,
		activityType,
		activity.ApplicationId,
		activity.DeploymentId,
		activity.EnvironmentId,
		activity.ConfigurationId).Scan(&createdAt)
	if err != nil {
		return nil, errors.Wrap(err, "preparing insert query")
	}

	return &proto.Empty{}, nil
}
