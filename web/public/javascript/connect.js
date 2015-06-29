var Codetainer;

function resize(term) {
  console.log("IN RESIZE");

  var div = document.getElementById("codetainer");

  var cell = createCell(div);
  console.log(cell)
  var size = getSize(div, cell);


  var x = size.cols
  var y = size.rows

  console.log("word", x,y, $('body').height(), $("body").width())

  Codetainer.Resize(x,y, function() {
    // term.resize(x, y);
  });
}

function getSize(element, cell) {
  var wSubs   = element.offsetWidth - element.clientWidth,
    w       = element.clientWidth - wSubs,

    hSubs   = element.offsetHeight - element.clientHeight,
    h       = element.clientHeight - hSubs,

    x       = cell.clientWidth,
    y       = cell.clientHeight,



    cols    = Math.max(Math.floor(w / x), 10),
    rows    = Math.max(Math.floor(h / y), 10),

    size    = {
    cols: cols,
    rows: rows
  };

  return size;
}

function createCell(element) {
  // $("#cell-size").remove();

  // var cell = $("<div id='cell-size'>").css({
    // position: "absolute",
    // top: "-10000px",
  // }).text("&nbsp").appendTo(element);

  // return cell.get(0);

  var cell            = document.createElement('div');

  cell.innerHTML      = '&nbsp';
  cell.style.position = 'absolute';
  cell.style.top      = '-1000px';

  element.appendChild(cell);

  return cell;
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
      // console.log("WTF", a,b,c,d)
    });
  },

  Build: function(container) {
    this.id = container

    Codetainer.Resize(80, 34, function() {
 
      var term = new Terminal({
        cols: 80,
        rows: 24,
        useStyle: false,
        screenKeys: true,
        cursorBlink: true
      });


      var div = document.getElementById("codetainer");

      term.open(div);

      var cell = createCell(div);
      var size = getSize(div, cell);


    //  var resizeTerm = resize.bind(null, term);
     // resizeTerm();
      //window.onresize = resizeTerm;


      console.log(term.element.offsetWidth, term.element.offsetHeight)

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


    })
  },

};

