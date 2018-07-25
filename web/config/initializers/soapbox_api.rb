Dir[Rails.root.join('lib/*_services_pb.rb')].each { |f| require f }

api_server = if ENV['SOAPBOX_API_SERVER'].blank?
               'localhost:9090'
             else
               ENV['SOAPBOX_API_SERVER']
             end
# TODO(paulsmith): consolidate API clients
$api_client = Soapbox::Applications::Stub.new(api_server, :this_channel_is_insecure)
$api_environment_client = Soapbox::Environments::Stub.new(api_server, :this_channel_is_insecure)
$api_configurations_client = Soapbox::Configurations::Stub.new(api_server, :this_channel_is_insecure)
$api_deployment_client = Soapbox::Deployments::Stub.new(api_server, :this_channel_is_insecure)
$api_user_client = Soapbox::Users::Stub.new(api_server, :this_channel_is_insecure)
$api_version_client = Soapbox::Version::Stub.new(api_server, :this_channel_is_insecure)
$api_activity_client = Soapbox::Activities::Stub.new(api_server, :this_channel_is_insecure)
