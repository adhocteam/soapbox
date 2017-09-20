require 'application_pb'

class DashboardController < ApplicationController

  before_action :get_activities

  helper_method :get_application_name
  helper_method :get_environment_name
  helper_method :get_user_name

  def index
    req = Soapbox::ListApplicationRequest.new(user_id: current_user.id)
    res = $api_client.list_applications(req)
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
    @app = $api_client.get_application(req)
    @app.name
  end

  def get_activities
    @activities = $api_activity_client
      .list_activities(Soapbox::Empty.new)
      .activities.sort_by {|a| a.created_at.seconds}.reverse
  end

  def get_user_name(user_id)
    req = Soapbox::GetUserRequest.new(id: user_id)
    @user_name = $api_user_client.get_user(req)
    @user_name.name
  end

  def get_environment_name(env_id)
    req = Soapbox::GetEnvironmentRequest.new(id: env_id)
    @env = $api_environment_client.get_environment(req)
    @env.name
  end

  def get_environments(app_id)
    req = Soapbox::ListEnvironmentRequest.new(application_id: app_id)
    $api_environment_client.list_environments(req).environments
  end

  def get_latest_deploy(app_id, env_id)
    req = Soapbox::GetLatestDeploymentRequest.new(application_id: app_id, environment_id: env_id)
    $api_deployment_client.get_latest_deployment(req)
  end
end
