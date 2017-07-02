require 'test_helper'

class SoapboxApplicationControllerTest < ActionDispatch::IntegrationTest
  test "should get index" do
    get soapbox_application_index_url
    assert_response :success
  end

end
