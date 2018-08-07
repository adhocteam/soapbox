class ApplicationController < ActionController::Base
  protect_from_forgery with: :exception
  before_action :require_login

  helper_method :current_user
  helper_method :refresh_user
  helper_method :octokit
  helper_method :get_metrics

  def require_login
    redirect_to login_user_path unless current_user && session[:login_token]
  end

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
    $api_client.users.get_user(req)
  end

  def user_metadata
    {
      metadata: {
        user_id: current_user.id.to_s,
        login_token: session[:login_token]
      }
    }
  end

  def get_metrics(app_id, metric_type)
    req = Soapbox::GetApplicationMetricsRequest.new(id: app_id, metric_type: metric_type)
    @metrics = []
    app_metrics = $api_client.applications.get_application_metrics(req, user_metadata)
    sorted_metrics = app_metrics.metrics.sort_by {|metric|
      Time.parse(metric.time)
    }
    sorted_metrics.each do |metric|
      m = [metric.time, metric.count]
      @metrics << m
    end

    @metrics
  end

  def octokit
    if current_user && current_user.github_oauth_access_token
      @octokit ||= Octokit::Client.new(access_token: current_user.github_oauth_access_token)
    end
  end
end
