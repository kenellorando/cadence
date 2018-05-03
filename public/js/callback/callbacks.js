// Callbacks.js: This file contains all callback classes used in theme.json
// This file must be loaded before theme.json

// Call this function to register a callback class as inheriting from the CallbackInterface class
// This is required to ensure compliance with the interfaces used by the theme engine.
function registerCallback(child) {// Save the current constructor name (so we can access it later)
    var name=child.prototype.constructor.name

    // Shim function: Call the parent constructor, then the child, in such a way that examining the constructor shows the original type name
    var shim=eval("(function() {\n\
       CallbackInterface.apply(this, arguments)\n\
       \n\
       "+name+".apply(this, arguments)\n\
    })")

    child.prototype=Object.create(CallbackInterface.prototype)
    child.prototype.constructor=shim
}
