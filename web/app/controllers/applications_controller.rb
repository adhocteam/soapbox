require 'application_pb'

class ApplicationsController < ApplicationController
  def index
    req = Soapbox::ListApplicationRequest.new
    res = $api_client.list_applications(req)
    if res.applications.count == 0
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
      req = Soapbox::CreateApplicationRequest.new(name: @form.name,
                                                  description: @form.description,
                                                  githubRepoURL: @form.github_repo_url,
                                                  type: type)
      $api_client.create_application(req)
      redirect_to applications_path
    else
      render :new
    end
  end

  def show
    req = Soapbox::GetApplicationRequest.new(id: params[:id].to_i)
    @app = $api_client.get_application(req).app
  end
end
