var Codetainer;

function getTextWidth(text, font) {
  // re-use canvas object for better performance
  var canvas = getTextWidth.canvas || (getTextWidth.canvas = document.createElement("canvas"));
  var context = canvas.getContext("2d");
  context.font = font;
  var metrics = context.measureText(text);
  return metrics.width;
};

function resize(term) {
  term.fit();
}

function getSize(element, cell) {
  var wSubs   = element.offsetWidth - element.clientWidth,
    w       = element.clientWidth - wSubs,

    hSubs   = element.offsetHeight - element.clientHeight,
    h       = element.clientHeight - hSubs,

    x       = cell.clientWidth,
    y       = cell.clientHeight,


    cols    = Math.max(Math.floor(w / getTextWidth("X", "11pt monospace")), 10),
    rows    = Math.max(Math.floor(h / y), 10),


    size    = {
    cols: cols,
    rows: rows
  };

  console.log("--- CELL --->", w, x, h, y)

  return size;
}

function createCell(element) {
  var cell            = document.createElement('div');

  cell.innerHTML      = '&nbsp';
  cell.style.position = 'absolute';
  cell.style.top      = '-1000px';
  cell.style["white-space"] = "nowrap";

  element.appendChild(cell);

  var s =  getTextWidth("X", "11pt monospace");
  console.log("#############", s, Math.floor(s));

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
    });
  },

  Build: function(container) {
    this.id = container

    var term = new Terminal({
      cols: 80,
      rows: 34,
      useStyle: true,
      screenKeys: true,
      cursorBlink: true
    });


    var div = document.getElementById("codetainer-terminal");
    term.open(div);

    Codetainer.term = term;

    var resizeTerm = resize.bind(null, Codetainer.term);
    resizeTerm();
    window.onresize = resizeTerm;


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
      websocket.send("\n");
      resizeTerm();
    }  

    function onClose(evt) { 
      term.write("Session terminated");
    }  

    function onMessage(evt) { 
      // console.log(evt);
      var data = evt.data.replace(/\+/g, '%20'); 
      data = decodeURIComponent(data)
      term.write(data);
    }  

    function onError(evt) { 
    }  


  },

};

Xterm.prototype.fit = function () {
  var self = this;
  var geometry = this.proposeGeometry();

  console.log("IN RESIZE WORD!!!!!!!!", geometry)

  self.resize(geometry.cols, geometry.rows);
  Codetainer.Resize(geometry.cols, geometry.rows, function() {
    self.resize(geometry.cols, geometry.rows);
  });
};


