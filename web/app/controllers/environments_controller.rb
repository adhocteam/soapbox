require 'environment_pb'

class EnvironmentsController < ApplicationController
  before_action :set_application, only: [:index, :create, :new, :show, :copy]

  def index
    @environments = $environments
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
      vars = []
      @form.environment_variables.each do |pair|
        name, value = pair
        vars << Soapbox::EnvironmentVariable.new(name: name, value: value)
      end
      env = Soapbox::Environment.new(application_id: params[:application_id].to_i, name: @form.name, vars: vars)
      $api_environment_client.create_environment(env)
      redirect_to application_environments_path
    else
      render :new
    end
  end

  def show
    @environment = get_environment(params[:id].to_i)
  end

  def destroy
    req = Soapbox::DestroyEnvironmentRequest.new(id: params[:id].to_i)
    $api_environment_client.destroy_environment(req)
    redirect_to application_environments_path
  end

  def copy
    env = get_environment(params[:id].to_i)
    names, values = [], []
    env.vars.each do |var|
      names << var.name
      values << var.value
    end
    env.name += " copy"
    @form = CreateEnvironmentForm.new({name: env.name, names: names, values: values})
    render :new
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
