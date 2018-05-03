// This file contains the specification for the parent 'skeleton' of the theme callback system.

// CallbackInterface constructor
// Takes the theme object to which this callback instance is attached
// The callback will not attempt to touch the callback field of the theme object.
// This is so that the callback can adapt it's behavior to the traits of the attached theme.
// Patterns which involve completely different behavior depending on the theme
//   (For example, a callback where all functions have no code outside a large switch of themekeys)
//   are very heavily discouraged.
// The intent of this is to allow a theme to, for example, have different behavior based on whether it's theme has a nightmode.

// In a theme object in JSON, there is a callback field. This shall take the name of a class to instantiate.
// If this field is omitted, the implementation will take behavior which is equivalent to using this base class.
// An instantiated theme has it's callback field as a constructed callback object, if it exists.
// Either way, the callback will be some object which implements the same interface as CallbackInterface -
//   For simplicity, callback types should inherit from this class.
function CallbackInterface(theme) {
    this.theme=theme
}

// Called when the theme is about to be loaded
// The return value for this function is checked:
//   If the return value is falsy, then the theme switch will occur.
//   If the return value is truthy, then it must be a valid theme object, indicating the theme to load instead.
// The single parameter to this function is the currently loaded theme object.
CallbackInterface.prototype.preLoad=function (currentTheme) {
    return false; // By default, theme load is permitted.
}

// Called when the theme has just been loaded.
// When this is called, some delay has passed since the document had its theme set.
// This delay is not guaranteed to be any particular value...
// But this function should be able to assume that the theme has been fully loaded, and the document is in a steady state with that theme.
// Return value is ignored. No parameters are passed.
CallbackInterface.prototype.postLoad=function () {}

// Called when the theme is about to be unloaded.
// When this is called, the theme will be unloaded just after the function returns.
// This function is responsible for canceling any pending timeouts that require the theme to be loaded.
// The return value of this function is ignored.
// The single passed parameter is the theme object which is about to be loaded.
//   (So that, if the theme to be loaded is known to share some resources, they don't need to be deallocated - the new callback's preLoad can use them)
CallbackInterface.prototype.preUnload=function (pendingTheme) {}

// Called just after the theme is unloaded.
// The same guarantees with regard to timing are made here as for postLoad.
// The return value is ignored. No parameters are passed.
// The callback may expect this instance to be unused after this call returns.
CallbackInterface.prototype.postUnload=function () {}

// Called before switching the theme into nightmode.
// This will be called before preUnload. It may set data to be used by postUnload later.
// The return value is ignored. No parameters are passed.
CallbackInterface.prototype.nightmodeSwitch=function () {}

// Called before switching the theme out of nightmode.
// Will, again, be called before preUnload, and may also set data for postUnload.
// Return value is ignored. No parameters are passed.
CallbackInterface.prototype.daymodeSwitch=function () {}
