require 'application_pb'

class DashboardController < ApplicationController
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

  def get_environments(app_id)
    req = Soapbox::ListEnvironmentRequest.new(application_id: app_id)
    $api_environment_client.list_environments(req).environments
  end

  def get_latest_deploy(app_id, env_id)
    req = Soapbox::GetLatestDeploymentRequest.new(application_id: app_id, environment_id: env_id)
    $api_deployment_client.get_latest_deployment(req)
  end
end
