class LoginUserForm
  include ActiveModel::Validations

  attr_reader :email, :password

  validates :email, presence: true
  validates :password, presence: true

  def initialize(attributes = {})
    @email = attributes[:email]
    @password = attributes[:password]
  end
end
