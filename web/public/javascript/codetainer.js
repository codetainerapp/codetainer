(function ($) {

  function Codetainer(element, options) {
    var self = this;

    self.options = options;
    self.$element = $(element);

    if (self.$element.data("url") !== undefined) {
      options.host = self.$element.data("url");
    }

    if (self.$element.data("container") !== undefined) {
      options.container = self.$element.data("container");
    }

    if (self.$element.data("containerHost") !== undefined) {
      options.containerHost = self.$element.data("host");
    }

    return self;
  };

  Codetainer.prototype.Build = function() {
    var self = this;

    var url = self.options.url + "/api/v1/codetainer/" +
    self.options.container + "/view";

    var iframe = "<iframe height='" + self.options.height + "'"+ 
      "width='" + self.options.width + "' title='Codetainer' scrolling='no' " +
      "frameborder='0' allowfullscreen='true' " +
      "allowtransparency='true' " +
      "src='"+url+"'>" +
      "style='box-shadow: 0px 0px 6px rgba(214,214,214,0.87);' " +
      "</iframe>";

    self.$element.html(iframe);
  };

  Codetainer.prototype.Fullscreen = function() {
    var self = this;
  };

  Codetainer.prototype.Resize = function(height, width) {
    var self = this;
  };

  Codetainer.prototype.Close = function(height, width) {
    var self = this;
  };

  $.fn.codetainer = function(options) {

    // This is the easiest way to have default options.
    var settings = $.extend({
      url: "https://localhost:3000",
      containerHost: "localhost:4500",
      width: "700px",
      height: "400px"
    }, options);

    $(this.selector).each(function(i, el) {
      var ct = new Codetainer(el, settings);

      $.data(el, "codetainer", ct);

      ct.Build();
    });

    return this;
  };

}(jQuery));
