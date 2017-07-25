require 'application_pb'
require 'deployment_pb'

class ApplicationsController < ApplicationController
  def index
    res = $api_client.list_applications(Soapbox::Empty.new)
    if res.applications.count.zero?
      redirect_to new_application_path
    else
      @applications = res.applications
    end
  end

  def new
    @form = CreateApplicationForm.new
  end

  def create
    @form = CreateApplicationForm.new(params[:application])
    if @form.valid?
      types = {
        'server' => Soapbox::ApplicationType::SERVER,
        'cronjob' => Soapbox::ApplicationType::CRONJOB
      }
      type = types[@form.type]
      app = Soapbox::Application.new(name: @form.name,
                                     description: @form.description,
                                     github_repo_url: @form.github_repo_url,
                                     type: type)
      app = $api_client.create_application(app)
      redirect_to application_path(app.id)
    else
      render :new
    end
  end

  def show
    req = Soapbox::GetApplicationRequest.new(id: params[:id].to_i)
    @app = $api_client.get_application(req)

    req = Soapbox::ListDeploymentRequest.new(application_id: params[:id].to_i)
    res = $api_deployment_client.list_deployments(req)
    @deployment = res.deployments.sort_by { |d| -Time.parse(d.created_at).to_i }.first
  end

  def confirm_delete
    req = Soapbox::GetApplicationRequest.new(id: params[:id].to_i)
    @app = $api_client.get_application(req)
  end

  def destroy
    req = Soapbox::DeleteApplicationRequest.new(id: params[:id].to_i)
    res = $api_client.delete_application(req)
  end
end
