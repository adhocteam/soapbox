Rails.application.routes.draw do
  get 'dashboard/index'

  resources :applications do
    get :confirm_delete, on: :member
    resources :environments do
      get :copy, on: :member
    end
    resources :deployments
  end

  root 'dashboard#index'
end
