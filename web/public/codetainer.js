window.docker = (function(docker) {
  docker.terminal = {
    startTerminalForContainer: function(host, container) {
      var term = new Terminal();
      term.open();

      var wsUri = "ws://" + 
        host + 
        "/v1.5/containers/" + 
        container + 
        "/attach/ws?logs=1&stderr=1&stdout=1&stream=1&stdin=1";

      wsUri = "ws://127.0.0.1:3000/api/v1/codetainer/" + container + 
        "/attach";
      console.log(wsUri);
      var websocket = new WebSocket(wsUri);
      websocket.onopen = function(evt) { onOpen(evt) };
      websocket.onclose = function(evt) { onClose(evt) };
      websocket.onmessage = function(evt) { onMessage(evt) };
      websocket.onerror = function(evt) { onError(evt) };

      term.on('data', function(data) {
        websocket.send(data);
      });

      function onOpen(evt) { 
        term.write("Session started");
      }  

      function onClose(evt) { 
        term.write("Session terminated");
      }  

      function onMessage(evt) { 
  console.log(evt);
        term.write(evt.data);
      }  

      function onError(evt) { 
      }  
    }
  };

  return docker;
})(window.docker || {});

$(function() {
  $("[data-docker-terminal-container]").each(function(i, el) {
    var container = $(el).data('docker-terminal-container');
    var host = $(el).data('docker-terminal-host');
    docker.terminal.startTerminalForContainer(host, container);
  });
});
