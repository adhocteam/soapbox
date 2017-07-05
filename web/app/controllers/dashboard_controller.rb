class DashboardController < ApplicationController
  def index
    stub = Application::Stub.new('localhost:9090', :this_channel_is_insecure)
    res = stub.list_applications(ListApplicationRequest.new)
    @num_applications = res.applications.count
  end
end
