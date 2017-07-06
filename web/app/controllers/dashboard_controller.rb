require 'application_pb'
require 'application_services_pb'

class DashboardController < ApplicationController
  def index
    stub = Soapbox::Applications::Stub.new('localhost:9090', :this_channel_is_insecure)
    res = stub.list_applications(Soapbox::ListApplicationRequest.new)
    @num_applications = res.applications.count
  end
end
