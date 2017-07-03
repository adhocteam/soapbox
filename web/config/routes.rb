Rails.application.routes.draw do
  get 'dashboard/index'

  resources :applications

  root 'dashboard#index'
end
