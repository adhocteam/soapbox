require 'application_pb'

class DashboardController < ApplicationController
  def index
    res = $api_client.list_applications(Soapbox::Empty.new)
    @num_applications = res.applications.count
    @num_deployments = res.applications.map(&:id).map do |app_id|
      req = Soapbox::ListDeploymentRequest.new(application_id: app_id)
      $api_deployment_client.list_deployments(req)
        .deployments
        .select { |dep| dep.state == "success" }
        .length
    end.sum
  end
end
