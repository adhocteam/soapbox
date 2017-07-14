# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: environment.proto

require 'google/protobuf'

require 'soapbox_pb'
Google::Protobuf::DescriptorPool.generated_pool.build do
  add_message "soapbox.ListEnvironmentRequest" do
    optional :application_id, :int32, 1
  end
  add_message "soapbox.ListEnvironmentResponse" do
    repeated :environments, :message, 1, "soapbox.Environment"
  end
  add_message "soapbox.GetEnvironmentRequest" do
    optional :id, :int32, 1
  end
  add_message "soapbox.Environment" do
    optional :id, :int32, 1
    optional :application_id, :int32, 2
    optional :name, :string, 3
    optional :slug, :string, 4
    repeated :vars, :message, 5, "soapbox.EnvironmentVariable"
    optional :created_at, :string, 6
  end
  add_message "soapbox.EnvironmentVariable" do
    optional :name, :string, 1
    optional :value, :string, 2
  end
  add_message "soapbox.DestroyEnvironmentRequest" do
    optional :id, :int32, 1
  end
  add_message "soapbox.CopyEnvironmentRequest" do
    optional :id, :int32, 1
  end
end

module Soapbox
  ListEnvironmentRequest = Google::Protobuf::DescriptorPool.generated_pool.lookup("soapbox.ListEnvironmentRequest").msgclass
  ListEnvironmentResponse = Google::Protobuf::DescriptorPool.generated_pool.lookup("soapbox.ListEnvironmentResponse").msgclass
  GetEnvironmentRequest = Google::Protobuf::DescriptorPool.generated_pool.lookup("soapbox.GetEnvironmentRequest").msgclass
  Environment = Google::Protobuf::DescriptorPool.generated_pool.lookup("soapbox.Environment").msgclass
  EnvironmentVariable = Google::Protobuf::DescriptorPool.generated_pool.lookup("soapbox.EnvironmentVariable").msgclass
  DestroyEnvironmentRequest = Google::Protobuf::DescriptorPool.generated_pool.lookup("soapbox.DestroyEnvironmentRequest").msgclass
  CopyEnvironmentRequest = Google::Protobuf::DescriptorPool.generated_pool.lookup("soapbox.CopyEnvironmentRequest").msgclass
end
