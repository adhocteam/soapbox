require 'application_pb'

# Dummy test data
EnvVar = Struct.new(:name, :value)
Environment = Struct.new(:id, :name, :environment_variables)
$environments = [
  Environment.new(1, "Prod", [
                    EnvVar.new("PORT", "8080"),
                    EnvVar.new("DB_CONNECTION", "pgsql://localhost/foobar"),
                    EnvVar.new("API_TOKEN", "a865dc85204a6d8352fa2206fa28c3d4")
                  ]),
  Environment.new(2, "Staging 1", [
                    EnvVar.new("PORT", "8181"),
                    EnvVar.new("DB_CONNECTION", "pgsql://localhost/quux"),
                    EnvVar.new("API_TOKEN", "b4da41f85c4c1d9287f2d8f1fa1e5428")
                  ])
]

class EnvironmentsController < ApplicationController
  before_action :set_application, only: [:index, :create, :new, :show]

  def index
    @environments = $environments
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
      req = Soapbox::CreateEnvironmentRequest.new(name: @form.name)
      $api_client.create_environment(req)
      redirect_to application_environments_path
    else
      render :new
    end
  end

  def show
    @environment = $environments[params[:id].to_i - 1]
    # TODO: uncomment below (and make it work) and remove fake data above
    # req = Soapbox::GetEnvironmentRequest.new(id: params[:id].to_i)
    # @environment = $api_client.get_environment(req).app
  end

  def set_application
    req = Soapbox::GetApplicationRequest.new(id: params[:application_id].to_i)
    @app = $api_client.get_application(req).app
  end
end
