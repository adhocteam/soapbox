Dir[Rails.root.join('lib/*_services_pb.rb')].each { |f| require f }

api_server = ENV['SOAPBOX_API_SERVER'] || 'localhost:9090'

$api_client = OpenStruct.new(
  applications: Soapbox::Applications::Stub.new(api_server, :this_channel_is_insecure),
  environments: Soapbox::Environments::Stub.new(api_server, :this_channel_is_insecure),
  configurations: Soapbox::Configurations::Stub.new(api_server, :this_channel_is_insecure),
  deployments: Soapbox::Deployments::Stub.new(api_server, :this_channel_is_insecure),
  users: Soapbox::Users::Stub.new(api_server, :this_channel_is_insecure),
  versions: Soapbox::Version::Stub.new(api_server, :this_channel_is_insecure),
  activities: Soapbox::Activities::Stub.new(api_server, :this_channel_is_insecure)
)
