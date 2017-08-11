class CreateConfigurationForm
  include ActiveModel::Validations

  attr_reader :config_vars

  validates :config_vars, presence: true

  def initialize(attributes = {})
    @config_vars = []
    (attributes[:names] || []).each_with_index do |name, i|
      value = attributes[:values][i]
      @config_vars << [name, value]
    end
  end
end
