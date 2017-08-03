Rails.application.routes.draw do
  get 'dashboard/index'

  resources :applications do
    resources :environments do
      get :copy, on: :member
    end
    resources :deployments
  end

  get '/about', to: 'about#index'

  root 'dashboard#index'
end
