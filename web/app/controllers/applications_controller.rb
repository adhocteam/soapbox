require 'application_pb'
require 'application_services_pb'

class ApplicationsController < ApplicationController
  def index
    req = ListApplicationRequest.new
    stub = Application::Stub.new('localhost:9090', :this_channel_is_insecure)
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
        'server' => CreateApplicationRequest::Type::SERVER,
        'cronjob' => CreateApplicationRequest::Type::CRONJOB
      }
      type = types[@form.type]
      req = CreateApplicationRequest.new(name: @form.name,
                                         description: @form.description,
                                         githubRepoURL: @form.github_repo_url,
                                         type: type)
      stub = Application::Stub.new('localhost:9090', :this_channel_is_insecure)
      stub.create_application(req)
      redirect_to applications_path
    else
      render :new
    end
  end

  def show
    stub = Application::Stub.new('localhost:9090', :this_channel_is_insecure)
    @app = stub.get_application(GetApplicationRequest.new(id: params[:id].to_i))
  end
end

