var Codetainer;

function resize(term) {
  console.log("IN RESIZE");

  // var x = $(".terminal").width();
  // var y = $('.terminal').height();

  var x = document.body.clientWidth / term.element.offsetWidth;
  var y = document.body.clientHeight / term.element.offsetHeight;

  x = x * term.cols | 0;
  y = y * term.rows | 0;

  // x = x * term.cols | 0;
  // y = y * term.rows | 0;

  console.log("word", x,y, $('body').height(), $("body").width())

  Codetainer.Resize(x,y, function() {
    term.resize(x, y);
  });
}

Codetainer = {
  id: "",

  Ajax: {
    Cache: {},

    Fetch: function(opts, callback, errback) {

      if (Codetainer.Ajax.Cache.hasOwnProperty(opts.url)) {
        Codetainer.Ajax.Cache[opts.url].abort();
      }

      var options = {
        dataType: "json",
        success: function(data) {
          delete Codetainer.Ajax.Cache[options.url]

          if (callback && typeof callback === "function") {
            return callback(data)
          } else {
            console.log(data);
          }
        },
        error: function(a, b, c) {
          delete Codetainer.Ajax.Cache[options.url]

          if (errback && typeof errback === "function") {
            return errback(a, b, c);
          } else {
            console.log(a, b, c)
          }
        }
      };

      $.extend(opts, options)

      Codetainer.Ajax.Cache[opts.url] = $.ajax(opts)
    },

    error: function(a, b, c) {
      console.log(a, b, c);
    }
  },

  Resize: function(x, y, callback) {
    console.log("Woprd?")
    Codetainer.Ajax.Fetch({
      url: "/api/v1/codetainer/" + Codetainer.id + "/tty",
      data: {
        height: y,
        width: x
      },
      dataType: "json",
      type: "post"
    }, function(data) {
      console.log("HJELLLLO", data);

      if (callback && typeof callback === "function") {
        return callback(data);
      }
    }, function(a,b,c,d) {
      console.log("WTF", a,b,c,d)

      if (callback && typeof callback === "function") {
        // return callback();
      }
    });
  },

  Build: function(container) {
    this.id = container

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

