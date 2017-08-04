require 'user_pb'

class UsersController < ApplicationController
  before_action :assign_user, only: %i[show omniauth]

  def new
    @form = CreateUserForm.new
  end

  def create
    @form = CreateUserForm.new(params[:user])
    if @form.valid?
      req = Soapbox::CreateUserRequest.new(name: @form.name, email: @form.email, password: @form.password)
      @user = $api_user_client.create_user(req)
      session[:current_user_email] = @user.email
      redirect_to profile_user_path
    else
      render :new
    end
  end

  def login
    @form = LoginUserForm.new
  end

  def attempt_login
    @form = LoginUserForm.new(params[:user])
    if @form.valid?
      req = Soapbox::LoginUserRequest.new(email: @form.email, password: @form.password)
      res = $api_user_client.login_user(req)
      if !res.error.blank?
        @form.errors.add(:password, :invalid)
        render :login
      else
        session[:current_user_email] = res.user.email
        redirect_to profile_user_path
      end
    else
      render :login
    end
  end

  def omniauth
    auth = request.env['omniauth.auth']
    @user.github_oauth_access_token = auth['credentials']['token']
    @user = $api_user_client.assign_github_omniauth_token_to_user(@user)
    refresh_user
    redirect_to profile_user_path
  end

  private

    def assign_user
      @user = current_user
    end
end
