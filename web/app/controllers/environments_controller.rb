require 'environment_pb'

class EnvironmentsController < ApplicationController
  before_action :set_application, only: [:index, :create, :new, :show]

  def index
    req = Soapbox::ListEnvironmentRequest.new(application_id: params[:application_id].to_i)
    res = $api_environment_client.list_environments(req)
    if res.environments.count == 0
      redirect_to new_application_environment_path
    else
      @environments = res.environments
    end
  end

  def new
    @form = CreateEnvironmentForm.new
  end

  def create
    @form = CreateEnvironmentForm.new(params[:environment])
    if @form.valid?
      env = Soapbox::Environment.new(application_id: params[:application_id].to_i, name: @form.name)
      $api_environment_client.create_environment(env)
      redirect_to application_environments_path
    else
      render :new
    end
  end

  def show
    env_id = params[:id].to_i
    @environment = get_environment(env_id)
    req = Soapbox::GetLatestConfigurationRequest.new(environment_id: env_id)
    @configuration = $api_configurations_client.get_latest_configuration(req)
  end

  def destroy
    req = Soapbox::DestroyEnvironmentRequest.new(id: params[:id].to_i)
    $api_environment_client.destroy_environment(req)
    redirect_to application_environments_path
  end

  private

  def set_application
    req = Soapbox::GetApplicationRequest.new(id: params[:application_id].to_i)
    @app = $api_client.get_application(req)
  end

  def get_environment(id)
    req = Soapbox::GetEnvironmentRequest.new(id: id)
    $api_environment_client.get_environment(req)
  end
end
