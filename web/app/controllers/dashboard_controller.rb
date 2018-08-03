require 'application_pb'

class DashboardController < ApplicationController
  before_action :get_activities
  helper_method :get_application_name
  helper_method :get_environment_name
  helper_method :get_user_name

  def index
    req = Soapbox::ListApplicationRequest.new(user_id: current_user.id)
    res = $api_client.applications.list_applications(req, user_metadata)
    apps = res.applications
    @num_applications = apps.count
    @num_deployments = 0
    apps.each do |app|
      envs = get_environments(app.id)
      envs.each do |env|
        begin
          latest_deploy = get_latest_deploy(app.id, env.id)
        rescue GRPC::NotFound
          latest_deploy = nil
        end
        if latest_deploy&.state == "success"
          @num_deployments += 1
        end
      end
    end
  end

  private

  def get_application_name(app_id)
    req = Soapbox::GetApplicationRequest.new(id: app_id)
    @app = $api_client.applications.get_application(req, user_metadata)
    @app.name
  end

  def get_activities
    @activities = $api_client.activities
      .list_activities(Soapbox::Empty.new, user_metadata)
      .activities.sort_by {|a| a.created_at.seconds}.reverse
  end

  def get_user_name(user_id)
    req = Soapbox::GetUserRequest.new(id: user_id)
    @user_name = $api_client.users.get_user(req, user_metadata)
    @user_name.name
  end

  def get_environment_name(env_id)
    req = Soapbox::GetEnvironmentRequest.new(id: env_id)
    @env = $api_client.environments.get_environment(req, user_metadata)
    @env.name
  end

  def get_environments(app_id)
    req = Soapbox::ListEnvironmentRequest.new(application_id: app_id)
    $api_client.environments.list_environments(req, user_metadata).environments
  end

  def get_latest_deploy(app_id, env_id)
    req = Soapbox::GetLatestDeploymentRequest.new(application_id: app_id, environment_id: env_id)
    $api_client.deployments.get_latest_deployment(req, user_metadata)
  end
end
