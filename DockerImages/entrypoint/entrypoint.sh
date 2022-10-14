# Generate SSH keys
cd /etc/ssh/ && ssh-keygen -A

# Create directory for privilege separation
mkdir -p /run/sshd

# Start OpenSSH service
service ssh start

# Switch user roles
su $user
