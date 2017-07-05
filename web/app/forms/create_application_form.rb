class CreateApplicationForm
  include ActiveModel::Validations

  attr_reader :name, :description, :github_repo_url, :type
  
  validates :name, presence: true
  validates :github_repo_url, presence: true, format: { with: /\Ahttps:\/\/github\.com\/.+\/.+\z/ }
  # TODO(paulsmith): get these values from the protobuf generated code
  validates :type, presence: true, inclusion: { :in => %w(server cronjob) }

  def initialize(attributes = {})
    @name = attributes[:name]
    @description = attributes[:description]
    @github_repo_url = attributes[:github_repo_url]
    @type = attributes[:type]
  end
end
