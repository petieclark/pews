// Nav scroll effect
const nav = document.getElementById('nav');
window.addEventListener('scroll', () => {
  nav.classList.toggle('scrolled', window.scrollY > 10);
});

// Mobile menu
const toggle = document.getElementById('mobileToggle');
const links = document.getElementById('navLinks');
toggle.addEventListener('click', () => {
  toggle.classList.toggle('active');
  links.classList.toggle('open');
});
links.querySelectorAll('a').forEach(a => a.addEventListener('click', () => {
  toggle.classList.remove('active');
  links.classList.remove('open');
}));

// Fade-in on scroll
const observer = new IntersectionObserver((entries) => {
  entries.forEach(e => { if (e.isIntersecting) { e.target.classList.add('visible'); observer.unobserve(e.target); }});
}, { threshold: 0.1, rootMargin: '0px 0px -40px 0px' });
document.querySelectorAll('.fade-up').forEach(el => observer.observe(el));

// Mailchimp subscribe via fetch (no-cors mode)
function submitForm(e, prefix) {
  e.preventDefault();
  const form = e.target;
  const msgEl = document.getElementById(prefix + 'Msg');
  const email = form.querySelector('input[name="EMAIL"]').value;
  const fname = form.querySelector('input[name="FNAME"]');
  if (!email) return false;

  msgEl.textContent = 'Submitting...';
  msgEl.className = prefix === 'hero' ? 'hero-form-msg' : 'cta-form-msg';

  const data = new FormData();
  data.append('u', '7f239d430d3572b980795a2ab');
  data.append('id', 'e665c7e29c');
  data.append('EMAIL', email);
  if (fname && fname.value) data.append('FNAME', fname.value);
  data.append('group[bdb8993f60][47f5ba21c0]', '1');

  fetch('https://warpapaya.us4.list-manage.com/subscribe/post', {
    method: 'POST',
    body: data,
    mode: 'no-cors'
  }).then(() => {
    // no-cors means we can't read the response, but the request went through
    msgEl.textContent = '🎉 You\'re on the list! Check your email to confirm.';
    msgEl.classList.add('success');
    form.reset();
  }).catch(() => {
    msgEl.textContent = 'Something went wrong. Please try again.';
    msgEl.classList.add('error');
  });

  return false;
}

// FAQ accordion
document.querySelectorAll('.faq-question').forEach(btn => {
  btn.addEventListener('click', () => {
    const item = btn.parentElement;
    const answer = item.querySelector('.faq-answer');
    const isOpen = item.classList.contains('open');
    
    // Close all
    document.querySelectorAll('.faq-item.open').forEach(openItem => {
      openItem.classList.remove('open');
      openItem.querySelector('.faq-answer').style.maxHeight = '0';
    });
    
    if (!isOpen) {
      item.classList.add('open');
      answer.style.maxHeight = answer.scrollHeight + 'px';
    }
  });
});
