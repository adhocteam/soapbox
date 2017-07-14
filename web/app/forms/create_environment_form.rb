class CreateEnvironmentForm
  include ActiveModel::Validations

  attr_reader :name, :environment_variables

  validates :name, presence: true

  def initialize(attributes = {})
    @name = attributes[:name]
    @environment_variables = []
    (attributes[:names] || []).each_with_index do |name, i|
      value = attributes[:values][i]
      @environment_variables << [name, value]
    end
  end
end
