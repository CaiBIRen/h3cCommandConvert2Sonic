#
# Regular cron jobs for the sonic-unis-framework package
#
0 4	* * *	root	[ -x /usr/bin/sonic-unis-framework_maintenance ] && /usr/bin/sonic-unis-framework_maintenance
