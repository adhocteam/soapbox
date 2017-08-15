Rails.application.routes.draw do
  get 'dashboard/index'

  resources :applications do
    resources :environments do
      get :copy, on: :member
      resources :configurations
    end
    resources :deployments
  end

  resource :user, only: %i[new create] do
    get :profile, on: :member, action: :show
    get :login
    post :login, to: 'users#attempt_login'
  end

  get '/auth/:provider/callback', to: 'users#omniauth'
  get '/about', to: 'about#index'

  root 'dashboard#index'
end
