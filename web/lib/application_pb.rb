# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: application.proto

require 'google/protobuf'

require 'soapbox_pb'
require 'google/protobuf/timestamp_pb'
Google::Protobuf::DescriptorPool.generated_pool.build do
  add_message "soapbox.Application" do
    optional :id, :int32, 1
    optional :user_id, :int32, 2
    optional :name, :string, 3
    optional :description, :string, 4
    optional :external_dns, :string, 5
    optional :github_repo_url, :string, 6
    optional :dockerfile_path, :string, 7
    optional :entrypoint_override, :string, 8
    optional :type, :enum, 9, "soapbox.ApplicationType"
    optional :created_at, :message, 10, "google.protobuf.Timestamp"
    optional :slug, :string, 11
    optional :internal_dns, :string, 12
    optional :creation_state, :enum, 13, "soapbox.CreationState"
    optional :aws_encryption_key_arn, :string, 14
  end
  add_message "soapbox.ListApplicationRequest" do
    optional :user_id, :int32, 1
  end
  add_message "soapbox.ListApplicationResponse" do
    repeated :applications, :message, 1, "soapbox.Application"
  end
  add_message "soapbox.GetApplicationRequest" do
    optional :id, :int32, 1
  end
  add_message "soapbox.ApplicationMetric" do
    optional :time, :string, 1
    optional :count, :int32, 2
  end
  add_message "soapbox.ApplicationMetricsResponse" do
    repeated :metrics, :message, 1, "soapbox.ApplicationMetric"
  end
  add_message "soapbox.GetApplicationMetricsRequest" do
    optional :id, :int32, 1
    optional :metric_type, :enum, 2, "soapbox.MetricType"
  end
  add_enum "soapbox.ApplicationType" do
    value :SERVER, 0
    value :CRONJOB, 1
  end
  add_enum "soapbox.MetricType" do
    value :REQUEST_COUNT, 0
    value :LATENCY, 1
    value :HTTP_5XX_COUNT, 2
    value :HTTP_4XX_COUNT, 3
    value :HTTP_2XX_COUNT, 4
  end
  add_enum "soapbox.CreationState" do
    value :CREATE_INFRASTRUCTURE_WAIT, 0
    value :SUCCEEDED, 1
    value :FAILED, 2
  end
end

module Soapbox
  Application = Google::Protobuf::DescriptorPool.generated_pool.lookup("soapbox.Application").msgclass
  ListApplicationRequest = Google::Protobuf::DescriptorPool.generated_pool.lookup("soapbox.ListApplicationRequest").msgclass
  ListApplicationResponse = Google::Protobuf::DescriptorPool.generated_pool.lookup("soapbox.ListApplicationResponse").msgclass
  GetApplicationRequest = Google::Protobuf::DescriptorPool.generated_pool.lookup("soapbox.GetApplicationRequest").msgclass
  ApplicationMetric = Google::Protobuf::DescriptorPool.generated_pool.lookup("soapbox.ApplicationMetric").msgclass
  ApplicationMetricsResponse = Google::Protobuf::DescriptorPool.generated_pool.lookup("soapbox.ApplicationMetricsResponse").msgclass
  GetApplicationMetricsRequest = Google::Protobuf::DescriptorPool.generated_pool.lookup("soapbox.GetApplicationMetricsRequest").msgclass
  ApplicationType = Google::Protobuf::DescriptorPool.generated_pool.lookup("soapbox.ApplicationType").enummodule
  MetricType = Google::Protobuf::DescriptorPool.generated_pool.lookup("soapbox.MetricType").enummodule
  CreationState = Google::Protobuf::DescriptorPool.generated_pool.lookup("soapbox.CreationState").enummodule
end
