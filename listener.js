// listener
// A queue listener

// builtin
var EventEmitter = require('events').EventEmitter,
  util = require('util');

// vendor
var _ = require('underscore');

/*
Example Usage

    var listener = new Listener(blocking_operation, 10); // 10 maximum callbacks at once
    listener.on('message', function (err, element, done) {
      if (err) {
        // do something
      } else {
        // do something else
      }
      done();
    })
      .start();

    // some time later
    listener.end();
*/

// -- Listener type --
// func - a function accepting a callback of error and element
// max_exec - A maximum number of elements executing
function Listener(func, max_out) {
  this._func = func;
  this._max = max_out || 0;
  this._out = 0;
  this._end = false;
  this._ended = false;
  this._listening = false;

  var self = this;

  this.on('done', function () {
    self._out = Math.max(self._out - 1, 0);

    if (self._end && self._out === 0 && !self._listenning) {
      self.emit('end'); // We are ready to end
    }

    self.start();
  });

  this.on('end', function () {
    this._ended = true;
  });
}

util.inherits(Listener, EventEmitter);

Listener.prototype.start = function () {
  if (this._listening || this._end) {
    // dont listen twice or listen if we've ended
    return;
  }
  this._listening = true;

  this._requeue();
  return this;
};

Listener.prototype.end = function () {
  this._ended = this._ended || this._out === 0;

  if (this._ended) {
    // Idempotency
    this.emit('end');
  }

  this._end = true;
  return this;
};

Listener.prototype.done = function () {
  this.emit('done');
};

Listener.prototype._create_done = function () {
  var self = this,
    called = false;
  return function () {
    if (!called) {
      self.emit('done');
      called = true;
    }
  };
};

Listener.prototype._iteration = function iteration () {
  var self = this;

  // Check for end or max out
  if (!self._end && (self._max === 0 || self._out < self._max)) {

    // Call the blocking function
    self._func(function (err, element) {
      if (err) {
        self.emit('error', err);
      } else if (element) {
        self._out++;
        self.emit('message', element, self._create_done());
      }

      // Requeue the iteration
      self._requeue();
    });
  } else {
    self._listening = false;

    if (self._end && self._out === 0) {
      self.emit('end'); // We are ready to end
    }
  }
};

Listener.prototype._requeue = function () {
  process.nextTick(_.bind(this._iteration, this));
};

module.exports = Listener;