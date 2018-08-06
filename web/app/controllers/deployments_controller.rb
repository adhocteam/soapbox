require 'deployment_pb'

class DeploymentsController < ApplicationController
  before_action :set_application, only: %i[index create new show]
  before_action :set_environments, only: %i[new create]
  before_action :find_commits, only: :new

  def index
    req = Soapbox::ListDeploymentRequest.new(application_id: params[:application_id].to_i)
    res = $api_client.deployments.list_deployments(req, user_metadata)
    if res.deployments.count == 0
      redirect_to new_application_deployment_path
    else
      @deployments = res.deployments.sort_by { |d| -d.created_at.seconds }
      @active_by_env = {}
      @deployments.each do |d|
        env = d.env.slug
        if d.state == 'success' && !@active_by_env.key?(env)
          @active_by_env[env] = d
        end
      end
    end
  end

  def new
    req = Soapbox::ListEnvironmentRequest.new(application_id: params[:application_id].to_i)
    res = $api_client.environments.list_environments(req, user_metadata)
    if res.environments.count == 0
      redirect_to new_application_environment_path
    else
      @form = CreateDeploymentForm.new
    end
  end

  def create
    @form = CreateDeploymentForm.new(params[:deployment])
    if @form.valid?
      env = Soapbox::Environment.new(id: @form.environment_id)
      app = Soapbox::Application.new(id: params[:application_id].to_i)
      req = Soapbox::Deployment.new(committish: @form.committish, application: app, env: env)
      res = $api_client.deployments.start_deployment(req, user_metadata)
      redirect_to application_deployments_path(application_id: params[:application_id].to_i)
    else
      render :new
    end
  end

  def show
    req = Soapbox::GetDeploymentStatusRequest.new(id: params[:id].to_i)
    res = $api_client.deployments.get_deployment_status(req, user_metadata)
    @state = res.state

    respond_to do |format|
      format.html
      format.json { render plain: @state }
    end
  end

  private

  def list_environments(application_id)
    req = Soapbox::ListEnvironmentRequest.new(application_id: application_id)
    $api_client.environments.list_environments(req, user_metadata).environments
  end

  def set_application
    req = Soapbox::GetApplicationRequest.new(id: params[:application_id].to_i)
    @app = $api_client.applications.get_application(req, user_metadata)
  end

  def set_environments
    @environments = list_environments(params[:application_id].to_i)
  end

  def find_commits
    repo_pat = %r{https://(www\.)?github\.com/([^/]+)/(.+)}
    if repo_pat.match(@app.github_repo_url) do |m|
         repo = "#{m[2]}/#{m[3]}"
         if repo.end_with?('.git')
           repo = repo[0..-5]
         end
         # TODO(paulsmith): allow user to override branch to fetch commits from
         @commits = octokit.commits(repo, 'master', per_page: 1000).map do |c|
           ["#{c.commit.message} (#{c.sha})", c.sha]
         end
       end
    else
      # TODO(paulsmith): warn/error invalid GitHub repo URL
      @commits = []
    end
  end
end
