require "rails_helper"

RSpec.describe ApplicationsController, type: :controller do
  describe "GET #show" do
    let(:app) {::Soapbox::Application.new(
          name: "test_app",
          github_repo_url: "https://ouath_data@github.com/bananas/shorts.git"
        )
    }
    let(:ts) { ::Google::Protobuf::Timestamp.new(seconds: Time.now.to_i) }
    let(:env) { ::Soapbox::Environment.new(name: "test_env") }
    let(:dep) { ::Soapbox::Deployment.new(
          id: 1,
          application: app,
          created_at: ts,
          env: env,
          committish: "abcdef0123456789"
        )
    }

    before do
      allow($api_client).to receive(:get_application) { app }
      allow(Soapbox::ListDeploymentRequest).to receive(:new) { nil }
      allow($api_deployment_client).to receive(:list_deployments) {
        ::Soapbox::ListDeploymentResponse.new(
          deployments: [dep]
        )
      }
    end

    context "github URL with oauth and .git" do
      it "responds correctly" do
        get :show, params: {id: 1}
        expect(response.success?)
      end

      it "sets the correct github url" do
        get :show, params: {id: 1}
        expect(@controller.instance_variable_get(:@github_url)).to eq "https://github.com/bananas/shorts"
      end
    end

    context "github URL without .git or oauth" do
      let(:app) {::Soapbox::Application.new(
            name: "test_app",
            github_repo_url: "https://github.com/bananas/shorts"
          )
      }

      it "responds correctly" do
        get :show, params: {id: 1}
        expect(response.success?)
      end

      it "sets the correct github url" do
        get :show, params: {id: 1}
        expect(@controller.instance_variable_get(:@github_url)).to eq "https://github.com/bananas/shorts"
      end
    end
  end
end
