(function(window) {
    "use strict";

    // this component requires that
    // the element passed in have two data attributes
    // data-id and data-state. Also, the element needs to have
    // a child span to display the deployment state.
    class DeploymentsApp {
      constructor(element, appId) {
        this.appId = appId;
        this.element = element;
        this.poll = this.poll.bind(this);
      }

      start() {
        this.poll(); // initial request
        this.poller = setInterval(this.poll, 1000);
      }

      stop() {
        clearInterval(this.poller);
      }

      poll() {
        const element = this.element;
        const appId = this.appId;
        const { state, id } = element.dataset;

        if (state === 'success') {
          return this.stop();
        }
        const req = new XMLHttpRequest();

        req.onreadystatechange = () => {
          if (req.readyState == XMLHttpRequest.DONE ) {
            if (req.status == 200) {
              const deploymentStatus = req.responseText;
              if (deploymentStatus === 'success') {
                element.classList.add('latest');
                element.classList.remove('active');
              }
              element.dataset.state = req.responseText;
              element.getElementsByTagName("span")[0].innerHTML = req.responseText;
            } else {
              console.log('error occurred while polling');
            }
          }
        };

        req.open("GET", `/applications/${appId}/deployments/${id}.json`, true);
        req.send();
      }
    }

  window.DeploymentsApp = DeploymentsApp;
})(window);
