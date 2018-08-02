require 'environment_pb'

class EnvironmentsController < ApplicationController
  before_action :set_application, only: [:index, :create, :new, :show]

  def index
    app_id = params[:application_id].to_i
    req = Soapbox::ListEnvironmentRequest.new(application_id: app_id)
    res = $api_client.environments.list_environments(req, user_metadata)
    if res.environments.count == 0
      redirect_to new_application_environment_path
    else
      @environments = []
      res.environments.each do |env|
        begin
          latest_deploy = get_latest_deploy(app_id, env.id)
        rescue GRPC::NotFound
          latest_deploy = nil
        end
        @environments << [env, latest_deploy]
      end
    end
  end

  def new
    @form = CreateEnvironmentForm.new
  end

  def create
    @form = CreateEnvironmentForm.new(params[:environment])
    if @form.valid?
      env = Soapbox::Environment.new(application_id: params[:application_id].to_i, name: @form.name)
      $api_client.environments.create_environment(env, user_metadata)
      redirect_to application_environments_path
    else
      render :new
    end
  end

  def show
    env_id = params[:id].to_i
    @environment = get_environment(env_id)
  end

  def destroy
    req = Soapbox::DestroyEnvironmentRequest.new(id: params[:id].to_i)
    $api_client.environments.destroy_environment(req, user_metadata)
    redirect_to application_environments_path
  end

  private

  def set_application
    req = Soapbox::GetApplicationRequest.new(id: params[:application_id].to_i)
    @app = $api_client.applications.get_application(req, user_metadata)
  end

  def get_environment(id)
    req = Soapbox::GetEnvironmentRequest.new(id: id)
    $api_client.environments.get_environment(req, user_metadata)
  end

  def get_latest_deploy(app_id, env_id)
    req = Soapbox::GetLatestDeploymentRequest.new(application_id: app_id, environment_id: env_id)
    $api_client.deployments.get_latest_deployment(req, user_metadata)
  end
end
