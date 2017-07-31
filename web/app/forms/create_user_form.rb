class CreateUserForm
  include ActiveModel::Validations

  attr_reader :name, :email, :password

  validates :name, presence: true
  validates :email, presence: true
  validates :password, presence: true, length: { minimum: 8 }

  def initialize(attributes = {})
    @name = attributes[:name]
    @email = attributes[:email]
    @password = attributes[:password]
  end
end
