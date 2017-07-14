class CreateDeploymentForm
  include ActiveModel::Validations

  attr_reader :committish, :environment_id

  validates :committish, presence: true
  validates :environment_id, presence: true

  def initialize(attributes = {})
    @committish = attributes[:committish]
    @environment_id = attributes[:environment_id].to_i
  end
end
2
