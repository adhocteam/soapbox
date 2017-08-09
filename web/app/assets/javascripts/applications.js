(function(window) {
    "use strict";

    // TODO: DRY up code with shared polling component
    class ApplicationsApp {
      constructor(element) {
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
        const { state, id } = element.dataset;

        if (state !== 'CREATE_INFRASTRUCTURE_WAIT') {
          return this.stop();
        }
        const req = new XMLHttpRequest();

        req.onreadystatechange = () => {
          if (req.readyState == XMLHttpRequest.DONE ) {
            if (req.status == 200) {
              const applicationStatus = req.responseText;
              if (applicationStatus === 'SUCCEEDED') {
                element.classList.add('latest');
                element.classList.remove('active');
              }
              if (applicationStatus === 'FAILED') {
                element.classList.add('complete');
              }
              const currentState = JSON.parse(req.responseText).creation_state
              element.dataset.state = currentState;
              element.getElementsByTagName("span")[0].innerHTML = currentState;
            } else {
              console.log('error occurred while polling');
            }
          }
        };

        req.open("GET", `/applications/${id}.json`, true);
        req.send();
      }
    }

  window.ApplicationsApp = ApplicationsApp;
})(window);
