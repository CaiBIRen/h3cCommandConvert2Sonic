Name:           ${AGENT_NAME}
Version:       	${AGENT_VERSION}
Release:       	${AGENT_TAG}%{?dist}
Summary:        sonic unis framework service
License:        GPL
Source0:        %{name}-%{version}.tar.gz

%description


%prep
rm -rf ${RPM_BUILD_ROOT}

%setup -q -n%{name}

%build
make clean
#make %{?_smp_mflags}
make

%define debug_package %{nil}

%install
mkdir -p ${RPM_BUILD_ROOT}/usr/local/bin/
cp sonic-unis-framework  ${RPM_BUILD_ROOT}/usr/local/bin/sonic-unis-framework


# 日志目录
mkdir -p ${RPM_BUILD_ROOT}/var/log/sonic-unis-framework

# 配置文件
mkdir -p ${RPM_BUILD_ROOT}/etc/sonic-unis-framework
mkdir -p ${RPM_BUILD_ROOT}/usr/lib/systemd/system
mkdir -p ${RPM_BUILD_ROOT}/opt/sonic-unis-framework

echo "install cp sonic-unis-framework.service"
cp -f sonic-unis-framework.service ${RPM_BUILD_ROOT}/opt/sonic-unis-framework/sonic-unis-framework.service
# cp -f etc/config.json ${RPM_BUILD_ROOT}/etc/sonic-unis-framework/config.json

%post

if [ $1 = 1 ]; then
    echo "new install service"
    echo "cp sonic-unis-framework.service"
	

	echo "new install config.json"
	
    cp -f /opt/sonic-unis-framework/sonic-unis-framework.service /usr/lib/systemd/system/sonic-unis-framework.service
    systemctl daemon-reload
# start service
    systemctl start ${AGENT_NAME}
    systemctl enable ${AGENT_NAME}
fi

if [ $1 = 2 ]; then
    echo "update service"
    if [  -z "`diff /opt/sonic-unis-framework/sonic-unis-framework.service /usr/lib/systemd/system/sonic-unis-framework.service`" ]; then
        echo "service is exist"
    else
        echo "create new service"
        cp -f /opt/sonic-unis-framework/sonic-unis-framework.service /usr/lib/systemd/system/sonic-unis-framework.service
        systemctl daemon-reload
    fi
    systemctl restart ${AGENT_NAME}
fi

%preun
if [ $1 = 2 ]; then
    systemctl restart ${AGENT_NAME}
fi
if [ $1 = 0 ]; then
    systemctl stop ${AGENT_NAME}
fi

%postun
if [ $1 = 0 ]; then
    echo "uninstall service"
    rm -rf /usr/lib/systemd/system/sonic-unis-framework.service
    systemctl daemon-reload
fi


%clean
rm -rf ${RPM_BUILD_ROOT}

%files
%defattr(-,root,root)
/usr/local/bin/sonic-unis-framework

/opt/sonic-unis-framework/sonic-unis-framework.service

# %config(noreplace) /etc/network-cvk-agent/config.json


%dir
/var/log/sonic-unis-framework
/usr/lib/systemd/system

%doc

%changelog
