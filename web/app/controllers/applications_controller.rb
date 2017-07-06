require 'application_pb'
require 'application_services_pb'

class ApplicationsController < ApplicationController
  def index
    req = Soapbox::ListApplicationRequest.new
    stub = Soapbox::Applications::Stub.new('localhost:9090', :this_channel_is_insecure)
    res = stub.list_applications(req)
    @applications = res.applications
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
      stub = Soapbox::Applications::Stub.new('localhost:9090', :this_channel_is_insecure)
      stub.create_application(req)
      redirect_to applications_path
    else
      render :new
    end
  end

  def show
    stub = Soapbox::Applications::Stub.new('localhost:9090', :this_channel_is_insecure)
    req = Soapbox::GetApplicationRequest.new(id: params[:id].to_i)
    @app = stub.get_application(req).app
  end
end
