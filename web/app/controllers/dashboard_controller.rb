require 'application_pb'

class DashboardController < ApplicationController
  def index
    res = $api_client.list_applications(Soapbox::Empty.new)
    @num_applications = res.applications.count
  end
end
