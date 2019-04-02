# Separate top-level Builddesc

When you want to create project that both builds separately and can be part of a
larger project, you can't have a CONFIG in your Builddesc file, and possibly
you also want different COMPONENTs. To solve this, at the top-level only, the
file `Builddesc.top` is preferred over Builddesc. You can put your CONFIG etc.
there. To also include Builddesc, add `.` as a component.

Example Builddesc.top:

```
CONFIG(
	flavors[dev release]
)

COMPONENT([
	.
])
```
