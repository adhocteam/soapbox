require 'application_pb'
require 'deployment_pb'

class ApplicationsController < ApplicationController
  before_action :find_repositories, only: %i[new create], if: :current_user

  def index
    req = Soapbox::ListApplicationRequest.new(user_id: current_user.id)
    res = $api_client.list_applications(req)
    if res.applications.count == 0
      redirect_to new_application_path
    else
      @applications = []
      apps = res.applications
      apps.each do |app|
        envs = get_environments(app.id)
        latest_deploy = nil
        envs.each do |env|
          begin
            deploy = get_latest_deploy(app.id, env.id)
            if latest_deploy.nil?
              latest_deploy = deploy
            elsif latest_deploy.created_at.seconds < deploy.created_at.seconds
              latest_deploy = deploy
            end
          rescue GRPC::NotFound
            next
          end
        end
        @applications << [app, latest_deploy]
      end
    end
  end

  def new
    @form = CreateApplicationForm.new(user_id: current_user.id)
  end

  def create
    @form = CreateApplicationForm.new(params[:application])
    if @form.valid?
      types = {
        'server' => Soapbox::ApplicationType::SERVER,
        'cronjob' => Soapbox::ApplicationType::CRONJOB
      }
      type = types[@form.type]
      app = Soapbox::Application.new(name: @form.name,
                                     description: @form.description,
                                     github_repo_url: @form.github_repo_url,
                                     type: type,
                                     user_id: current_user.id)
      app = $api_client.create_application(app)
      redirect_to application_path(app.id)
    else
      render :new
    end
  end

  def show
    req = Soapbox::GetApplicationRequest.new(id: params[:id].to_i)
    @app = $api_client.get_application(req)

    req = Soapbox::ListDeploymentRequest.new(application_id: params[:id].to_i)
    res = $api_deployment_client.list_deployments(req)
    @deployment = res.deployments.sort_by { |d| -d.created_at.seconds }.first

    # strip the oauth token from the URL if present and remove trailing `.git`
    @github_url = @app.github_repo_url.gsub(%r{https://.*@}, 'https://').gsub(/\.git$/, '')

    respond_to do |format|
      format.html
      format.json { render json: @app.as_json }
    end
  end

  def destroy
    req = Soapbox::GetApplicationRequest.new(id: params[:id].to_i)
    @app = $api_client.get_application(req)
    $api_client.delete_application(@app)
    redirect_to application_path(@app.id)
  end

  private

  def find_repositories
    @repos = octokit.repositories(nil, sort: :pushed, per_page: 1000).map do |repo|
      if repo[:private]
        [repo[:clone_url], repo[:clone_url].gsub('https://', "https://#{current_user.github_oauth_access_token}@")]
      else
        repo[:clone_url]
      end
    end
  end

  def get_environments(app_id)
    req = Soapbox::ListEnvironmentRequest.new(application_id: app_id)
    $api_environment_client.list_environments(req).environments
  end

  def get_latest_deploy(app_id, env_id)
    req = Soapbox::GetLatestDeploymentRequest.new(application_id: app_id, environment_id: env_id)
    $api_deployment_client.get_latest_deployment(req)
  end
end
