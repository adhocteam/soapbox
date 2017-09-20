package soapboxd

import (
	"database/sql"
	"time"

	"github.com/adhocteam/soapbox/models"
	"github.com/adhocteam/soapbox/proto"
	pb "github.com/adhocteam/soapbox/proto"
	gpb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

var activityTypeToModel = map[pb.ActivityType]models.ActivityType{
	pb.ActivityType_APPLICATION_CREATED:   models.ActivityTypeApplicationCreated,
	pb.ActivityType_DEPLOYMENT_STARTED:    models.ActivityTypeDeploymentStarted,
	pb.ActivityType_DEPLOYMENT_SUCCESS:    models.ActivityTypeDeploymentSuccess,
	pb.ActivityType_DEPLOYMENT_FAILURE:    models.ActivityTypeDeploymentFailure,
	pb.ActivityType_ENVIRONMENT_CREATED:   models.ActivityTypeEnvironmentCreated,
	pb.ActivityType_ENVIRONMENT_DESTROYED: models.ActivityTypeEnvironmentDestroyed,
}

func (s *server) AddActivity(ctx context.Context, activity *pb.Activity) (*pb.Empty, error) {
	query := `
	INSERT INTO activities (user_id, activity, application_id, deployment_id, environment_id)
	VALUES ($1, $2, $3, $4, $5)
	`
	params := []interface{}{
		activity.UserId,
		activityTypeToModel[activity.Type],
		activity.ApplicationId,
		activity.DeploymentId,
		activity.EnvironmentId,
	}
	_, err := s.db.Exec(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "inserting into activities")
	}

	return &proto.Empty{}, nil
}

func (s *server) ListActivities(ctx context.Context, _ *pb.Empty) (*pb.ListActivitiesResponse, error) {
	translations := map[models.ActivityType]pb.ActivityType{
		models.ActivityTypeApplicationCreated:   pb.ActivityType_APPLICATION_CREATED,
		models.ActivityTypeDeploymentStarted:    pb.ActivityType_DEPLOYMENT_STARTED,
		models.ActivityTypeDeploymentSuccess:    pb.ActivityType_DEPLOYMENT_SUCCESS,
		models.ActivityTypeDeploymentFailure:    pb.ActivityType_DEPLOYMENT_FAILURE,
		models.ActivityTypeEnvironmentCreated:   pb.ActivityType_ENVIRONMENT_CREATED,
		models.ActivityTypeEnvironmentDestroyed: pb.ActivityType_ENVIRONMENT_DESTROYED,
	}

	const query = `
		SELECT id, user_id, activity, application_id, deployment_id,
		environment_id, created_at
		FROM activities
		ORDER BY created_at ASC
		LIMIT 50`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, errors.Wrap(err, "listing activites")
	}

	var activities []*pb.Activity

	for rows.Next() {
		activity := &pb.Activity{
			CreatedAt: new(gpb.Timestamp),
		}
		var (
			aType         models.ActivityType
			applicationId sql.NullInt64
			deploymentId  sql.NullInt64
			environmentId sql.NullInt64
		)
		var createdAt time.Time
		dest := []interface{}{
			&activity.Id,
			&activity.UserId,
			&aType,
			&applicationId,
			&deploymentId,
			&environmentId,
			&createdAt,
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, errors.Wrap(err, "scanning db row")
		}
		if applicationId.Valid {
			activity.ApplicationId = int32(applicationId.Int64)
		}
		if deploymentId.Valid {
			activity.DeploymentId = int32(deploymentId.Int64)
		}
		if environmentId.Valid {
			activity.EnvironmentId = int32(environmentId.Int64)
		}
		activity.Type = translations[aType]
		setPbTimestamp(activity.CreatedAt, createdAt)
		activities = append(activities, activity)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterating over db rows")
	}

	return &pb.ListActivitiesResponse{Activities: activities}, nil
}

func (s *server) AddApplicationActivity(ctx context.Context, applicationId int32, userId int32) error {
	query := `
	INSERT INTO activities (user_id, activity, application_id)
	VALUES ($1, $2, $3)
	`
	_, err := s.db.Exec(query,
		userId,
		models.ActivityTypeApplicationCreated,
		applicationId,
	)
	if err != nil {
		return errors.Wrap(err, "error adding application activity")
	}

	return nil
}

func (s *server) AddDeploymentActivity(ctx context.Context, activityType pb.ActivityType, d *deployState) error {
	query := `
	INSERT INTO activities (user_id, activity, application_id, deployment_id)
	VALUES ($1, $2, $3, $4)
	`
	_, err := s.db.Exec(query,
		d.userID,
		activityTypeToModel[activityType],
		d.app.id,
		d.id,
	)
	if err != nil {
		return errors.Wrap(err, "error adding deployment activity")
	}

	return nil
}

func (s *server) AddCreateEnvironmentActivity(ctx context.Context, env *pb.Environment) error {
	application, err := s.GetApplication(ctx, &pb.GetApplicationRequest{Id: env.GetApplicationId()})
	if err != nil {
		return errors.Wrap(err, "error adding environment activity")
	}

	query := `
	INSERT INTO activities (user_id, activity, environment_id)
	VALUES ($1, $2, $3)
	`
	_, err = s.db.Exec(query,
		application.GetUserId(),
		models.ActivityTypeEnvironmentCreated,
		env.GetId(),
	)
	if err != nil {
		return errors.Wrap(err, "error adding environment activity")
	}
	return nil
}

func (s *server) ListApplicationActivities(ctx context.Context, app *pb.GetApplicationRequest) (*pb.ListActivitiesResponse, error) {
	// XXX TODO
	var activities []*pb.Activity
	return &pb.ListActivitiesResponse{Activities: activities}, nil
}

func (s *server) ListDeploymentActivities(ctx context.Context, app *pb.GetDeploymentRequest) (*pb.ListActivitiesResponse, error) {
	// XXX TODO
	var activities []*pb.Activity
	return &pb.ListActivitiesResponse{Activities: activities}, nil
}
