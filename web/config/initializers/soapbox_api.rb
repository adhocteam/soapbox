require 'application_services_pb'
require 'environment_services_pb'

api_server = if ENV['SOAPBOX_API_SERVER'].blank?
               'localhost:9090'
             else
               ENV['SOAPBOX_API_SERVER']
             end
$api_client = Soapbox::Applications::Stub.new(api_server, :this_channel_is_insecure)
# TODO(paulsmith): consolidate API clients
$api_environment_client = Soapbox::Environments::Stub.new(api_server, :this_channel_is_insecure)
