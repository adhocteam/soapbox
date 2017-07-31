Rails.application.routes.draw do
  get 'dashboard/index'

  resources :applications do
    resources :environments do
      get :copy, on: :member
    end
    resources :deployments
  end

  resource :user, only: %i[new create show] do
    get :login
    post :login, to: 'users#attempt_login'
  end

  get '/auth/:provider/callback', to: 'users#omniauth'

  root 'dashboard#index'
end
