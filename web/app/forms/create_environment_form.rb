class CreateEnvironmentForm
  include ActiveModel::Validations

  attr_reader :name, :productionFlag

  validates :name, presence: true
  validates :productionFlag, presence: true

  def initialize(attributes = {})
    @name = attributes[:name]
    @productionFlag = attributes[:productionFlag]
  end
end
