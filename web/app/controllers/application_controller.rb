class ApplicationController < ActionController::Base
  protect_from_forgery with: :exception
  helper_method :current_user
  helper_method :refresh_user

  # Finds the User with the ID stored in the session with the key
  # :current_user_email
  def current_user
    @_current_user ||= session[:current_user_email] &&
                       get_user(session[:current_user_email])
  end

  def refresh_user
    @_current_user = session[:current_user_email] &&
                     get_user(session[:current_user_email])
  end

  def get_user(email)
    req = Soapbox::GetUserRequest.new(email: email)
    $api_user_client.get_user(req)
  end

  def authorize!
    redirect_to root_path unless current_user
  end
end
