require 'application_pb'

class EnvironmentsController < ApplicationController
  before_action :set_application, only: [:index, :create, :new]

  def index
    res = OpenStruct.new
    res.environments = []
    res.environments[0] = OpenStruct.new
    res.environments[0].name = "Prod"
    res.environments[0].productionFlag = true
    res.environments[1] = OpenStruct.new
    res.environments[1].name = "Staging 1"
    res.environments[1].productionFlag = false
    @environments = res.environments
    # TODO: uncomment below (and make it work) and remove fake data above
    # req = Soapbox::ListEnvironmentRequest.new
    # res = $api_client.list_environments(req)
    # if res.environments.count = 0
    #   redirect_to new_application_environment_path
    # else
    #   @environments = res.environments
    # end
  end

  def new
    @form = CreateEnvironmentForm.new
  end

  def create
    @form = CreateEnvironmentForm.new(params[:environment])
    if @form.valid?
      # TODO: make this work
      req = Soapbox::CreateEnvironmentRequest.new(name: @form.name,
                                                  productionFlag: @form.productionFlag)
      $api_client.create_environment(req)
      redirect_to application_environments_path
    else
      render :new
    end
  end

  def show
    environment = OpenStruct.new
    environment.environment_variables = []
    environment.environment_variables[0] = OpenStruct.new
    environment.environment_variables[0].key = "A key"
    environment.environment_variables[0].value = "A val"
    environment.environment_variables[0].type = "a type"
    environment.environment_variables[0].rate = 100
    environment.name = "Prod"
    environment.productionFlag = true
    environment.createdAt = "2017-07-11T12:33:45"
    @environment = environment
    @environment_variables = environment.environment_variables
    # TODO: uncomment below (and make it work) and remove fake data above
    # req = Soapbox::GetEnvironmentRequest.new(id: params[:id].to_i)
    # @environment = $api_client.get_environment(req).app
  end

  def set_application
    req = Soapbox::GetApplicationRequest.new(id: params[:application_id].to_i)
    @app = $api_client.get_application(req).app
  end
end
