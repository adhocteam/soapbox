Rails.application.routes.draw do
  get 'dashboard/index'

  resources :applications do
    resources :environments
  end

  root 'dashboard#index'
end
