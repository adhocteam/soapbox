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
  end

  def create
    types = {
      "server" => CreateApplicationRequest::Type::SERVER,
      "cronjob" => CreateApplicationRequest::Type::CRONJOB
    }
    type = types[params[:application][:type]]
    req = CreateApplicationRequest.new(:name => params[:application][:name],
                                                :description => params[:application][:description],
                                                :githubRepoURL => params[:application][:github_repo_url],
                                               :type => type)
    stub = Application::Stub.new('localhost:9090', :this_channel_is_insecure)
    stub.create_application(req)
    redirect_to applications_path
  end
end
