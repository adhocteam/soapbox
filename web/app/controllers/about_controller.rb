require 'version_pb'

class AboutController < ApplicationController
  def index
    @version = $api_version_client.get_version(Soapbox::Empty.new)
  end
end
