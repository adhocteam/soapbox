(function(window) {
    "use strict";

    class DeploymentsApp {
      constructor(element) {
        this.element = element;
        this.poll = this.poll.bind(this);
      }

      start() {
        this.poller = setInterval(this.poll, 3000);
      }

      stop() {
        clearInterval(this.poller);
      }

      poll() {
        const element = this.element;
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
                element.classList.add('complete');
              }
              element.classList.remove('active');
              element.getElementsByTagName("span")[0].innerHTML = req.responseText;
            } else {
              console.log('error occurred while polling');
            }
           }
         };

        req.open("GET", `deployments/${id}.json`, true);
        req.send();
      }
    }

  window.DeploymentsApp = DeploymentsApp;
})(window);
