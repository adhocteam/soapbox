package soapboxd

import (
	"time"

	"github.com/adhocteam/soapbox/models"
	"github.com/adhocteam/soapbox/proto"
	pb "github.com/adhocteam/soapbox/proto"
	gpb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// return value if present, nil otherwise. Is there a better golang way to type this?
func unwrap(value *pb.MaybeInt) interface{} {
	if value != nil && value.Present {
		return value.Value
	}
	return nil
}

func (s *server) AddActivity(ctx context.Context, activity *pb.Activity) (*pb.Empty, error) {
	translations := map[pb.ActivityType]models.ActivityType{
		pb.ActivityType_APPLICATION_CREATED: models.ActivityTypeApplicationCreated,
		pb.ActivityType_DEPLOYMENT_STARTED:  models.ActivityTypeDeploymentStarted,
		pb.ActivityType_DEPLOYMENT_SUCCESS:  models.ActivityTypeDeploymentSuccess,
		pb.ActivityType_DEPLOYMENT_FAILURE:  models.ActivityTypeDeploymentFailure,
		pb.ActivityType_ENVIRONMENT_CREATED: models.ActivityTypeEnvironmentCreated,
	}

	userID := unwrap(activity.UserId)
	applicationID := unwrap(activity.ApplicationId)
	deploymentID := unwrap(activity.DeploymentId)
	environmentID := unwrap(activity.EnvironmentId)

	insertQuery := `
	INSERT INTO activities (user_id, activity, application_id, deployment_id, environment_id)
	VALUES ($1, $2, $3, $4, $5)
	`
	var activityType = translations[activity.Type]
	_, err := s.db.Exec(insertQuery,
		userID,
		activityType,
		applicationID,
		deploymentID,
		environmentID)
	if err != nil {
		return nil, errors.Wrap(err, "inserting into activities")
	}

	return &proto.Empty{}, nil
}

func (s *server) ListActivities(ctx context.Context, _ *pb.Empty) (*pb.ListActivitiesResponse, error) {
	const query = `
		SELECT id, user_id, application_id, deployment_id,
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

		var createdAt time.Time
		dest := []interface{}{
			&activity.Id,
			&activity.UserId,
			&activity.ApplicationId,
			&activity.DeploymentId,
			&activity.EnvironmentId,
			&createdAt,
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, errors.Wrap(err, "scanning db row")
		}

		setPbTimestamp(activity.CreatedAt, createdAt)
		activities = append(activities, activity)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterating over db rows")
	}

	return &pb.ListActivitiesResponse{Activities: activities}, nil
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
