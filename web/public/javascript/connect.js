var Codetainer;

var TabControl;
function TabControl(options) {
  var self = this;
  self.options = options;
  self.tabs = [];

  return this;
}

TabControl.prototype.Render = function() {
  var self = this;

  $(".codetainer-header").append(Codetainer.Templates.tabs({}));

  return self;
}

TabControl.prototype.Add = function(data) {
  var self = this;

  data.position = 0

  if (self.tabs.length > 0) {
    data.position = self.tabs[self.tabs.length-1].options.position + 1;
  } 

  var tab = new Tab(data);
  self.tabs.push(tab);

  tab.Render();

  return tab;
}

TabControl.prototype.Remove = function(tab) {
  var self = this;

  tab.Remove();

  for (item in Codetainer.Tabs.tabs) {
    var t = Codetainer.Tabs.tabs[item];

    console.log(t.tabId, tab.tabId)

    if (t.tabId === tab.tabId) {
      var newindex = item - 1;

      if (item <= 0) {
        self.tabs[0].Show();
      } else {
        if (!self.tabs[0].open) {
          self.tabs[newindex].Show();
        }
      }
      self.tabs.splice(item, 1)
      console.log(self.tabs)
      return self;
    }
  }

  return self;
}

var Tab;
function Tab(options) {
  var self = this;
  self.options = options;
  self.tabId = Codetainer.id + "-" + self.options.position + "-tab";
  self.contentId = self.tabId + "-content";

  self.open = false;
  return this;
}

Tab.prototype.Render = function() {
  var self = this;

  var content = Codetainer.Templates.content(self);
  var tab = Codetainer.Templates.newTab(self);
  $(".codetainer-tabs .nav").append(tab);
  $(".codetainer-content").append(content);

  $("#" + self.tabId).on("click", "", function(e) {
    e.preventDefault()

    if (!self.open) {
      self.Show();
    }
  });

  $("#" + self.tabId + "-close").on("click", "", function(e) {
    e.preventDefault()
    e.stopPropagation();

    Codetainer.Tabs.Remove(self)
  });

  return self;
}

Tab.prototype.Show = function() {
  var self = this;
  self.open = true;

  for (tab in Codetainer.Tabs.tabs) {
    var t = Codetainer.Tabs.tabs[tab];

    if (t.tabId !== self.tabId) {
      t.Hide();
    }
  }

  // $(".codetainer-tabs li").removeClass("active");
  $("#" + self.tabId).addClass("active");
  $("#" + self.contentId).show();

  return self;
}

Tab.prototype.Hide = function() {
  var self = this;
  self.open = false;
  $("#" + self.tabId).removeClass("active");
  $("#" + self.contentId).hide();
  return self;
}

Tab.prototype.Remove = function() {
  var self = this;
  self.open = false;
  $("#" + self.tabId + "-close").off();
  $("#" + self.tabId).off().remove();
  $("#" + self.contentId).remove();
  return self;
}


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

  return cell;
}

Codetainer = {
  id: "",

  Templates: {
    Init: function() {
      var self = this;

      self.tabs = Handlebars.compile($("#codetainer-tab-control").html());
      self.newTab = Handlebars.compile($("#codetainer-tab-new").html());
      self.content = Handlebars.compile($("#codetainer-content").html());
      self.sidebar = Handlebars.compile($("#codetainer-sidebar").html());

      Handlebars.registerHelper ('truncate', function(str, len) {
        if (str.length > len) {
          var new_str = str.substr (0, len + 1);

          while (new_str.length) {
            var ch = new_str.substr (-1);
            new_str = new_str.substr (0, -1);

            if (ch == ' ') {
              break;
            }
          }

          if (new_str == '') {
            new_str = str.substr (0, len);
          }

          return new Handlebars.SafeString (new_str + '...');
        }
        return str;
      });
    }
  },

  Tabs: null,

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

      if (callback && typeof callback === "function") {
        return callback(data);
      }
    }, function(a,b,c,d) {
      console.log("WTF", a,b,c,d)
    });
  },

  Build: function(container) {
    this.id = container

    Codetainer.Templates.Init()

    var term = new Terminal({
      cols: 80,
      rows: 34,
      useStyle: true,
      screenKeys: true,
      cursorBlink: true
    });

    Codetainer.Tabs = new TabControl();
    Codetainer.Tabs.Render().Add({
      name: "Terminal",
      terminal: true,
      active: true,
      content: function() {
        return Codetainer.Templates.terminal({});
      }
    }).Show();

    var div = document.getElementById("codetainer-terminal");
    term.open(div);

    Codetainer.term = term;

    var resizeTerm = resize.bind(null, Codetainer.term);
    resizeTerm();
    window.onresize = resizeTerm;


    var host = location.hostname + ":" + location.port;
    var wsUri = "ws://" + host + "/api/v1/codetainer/" + container + 
    "/attach";

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
      var data = evt.data.replace(/\+/g, '%20'); 
      data = decodeURIComponent(data)
      term.write(data);
    }  

    function onError(evt) { 
    }  

    $(".create-new-folder").on("click", function(e) {
      e.preventDefault();
      Codetainer.Tabs.Add({
        name: Blahtest(),
        content: function() {
          return Blahtest();
        }
      })
    });
  },
};

Xterm.prototype.fit = function () {
  var self = this;
  var geometry = this.proposeGeometry();

  self.resize(geometry.cols, geometry.rows);
  Codetainer.Resize(geometry.cols, geometry.rows, function() {
    self.resize(geometry.cols, geometry.rows);
  });
};

function Blahtest()
{
    var text = "";
    var possible = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";

    for( var i=0; i < 10; i++ )
        text += possible.charAt(Math.floor(Math.random() * possible.length));

    return text;
}


