class CreateEnvironmentForm
  include ActiveModel::Validations

  attr_reader :name

  validates :name, presence: true

  def initialize(attributes = {})
    @name = attributes[:name]
  end
end
