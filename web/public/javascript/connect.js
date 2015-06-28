var Codetainer;

function resize(term) {
  console.log("IN RESIZE");

  var x = document.body.clientWidth / term.element.offsetWidth;
  var y = document.body.clientHeight / term.element.offsetHeight;

  console.log("word", x,y, $('body').height(), $("body").width())

  x = x * term.cols | 0;
  y = y * term.rows | 0;

  term.resize(x, y);
}

Codetainer = {

  Build: function(container) {

    var term = new Terminal({});

    term.open({
      cols: 80,
      rows: 34,
      useStyle: true,
      screenKeys: true,
      cursorBlink: true
    });

    var resizeTerm = resize.bind(null, term);
    resizeTerm();
    setTimeout(resizeTerm, 1000);
    window.onresize = resizeTerm;


    console.log(term.element.offsetWidth, term.element.offsetHeight)

    // term.resize()

    var wsUri = "ws://127.0.0.1:3000/api/v1/codetainer/" + container + 
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
      resizeTerm();
    }  

    function onClose(evt) { 
      term.write("Session terminated");
    }  

    function onMessage(evt) { 
      // console.log(evt);
      term.write(evt.data);
    }  

    function onError(evt) { 
    }  
  },

};

