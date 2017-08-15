require 'configuration_pb'

class ConfigurationsController < ApplicationController
  before_action :set_context, only: [:index, :create]

  def index
    begin
      @configuration = get_latest_config(@environment)
      @form = form_from_config(@configuration)
    rescue GRPC::NotFound
      @configuration = nil
      @form = CreateConfigurationForm.new
    end
  end

  def create
    @form = CreateConfigurationForm.new(params[:configuration])
    if @form.valid?
      vars = []
      @form.config_vars.each do |pair|
        name, value = pair
        vars << Soapbox::ConfigVar.new(name: name, value: value)
      end
      puts vars
      env_id = params[:environment_id].to_i
      req = Soapbox::CreateConfigurationRequest.new(environment_id: env_id, config_vars: vars)
      $api_configurations_client.create_configuration(req)
      redirect_to application_environment_path(id: env_id)
    else
      render :index
    end
  end

  private

  def set_context
    req = Soapbox::GetApplicationRequest.new(id: params[:application_id].to_i)
    @app = $api_client.get_application(req)
    req = Soapbox::GetEnvironmentRequest.new(id: params[:environment_id].to_i)
    @environment = $api_environment_client.get_environment(req)
  end

  def get_latest_config(env)
    req = Soapbox::GetLatestConfigurationRequest.new(environment_id: env.id)
    $api_configurations_client.get_latest_configuration(req)
  end

  def form_from_config(config)
    names, values = [], []
    config.config_vars.each_with_index do |var, i|
      names[i] = var.name
      values[i] = var.value
    end
    CreateConfigurationForm.new({ names: names, values: values })
  end
end
