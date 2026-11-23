import { profileUpdateBody } from './profile-update-body';

describe('profileUpdateBody', () => {
  const storedUser = {
    id: 'user-1',
    name: 'jane_doe',
    title: 'Jane Doe',
    email: 'jane@raystack.org',
    avatar: 'data:image/png;base64,old',
    metadata: { onboarded: true, team: 'core' }
  };

  it('leaves out metadata when only the avatar changes', () => {
    const formValues = { ...storedUser, avatar: 'data:image/png;base64,new' };

    const body = profileUpdateBody(formValues, storedUser.name);

    expect(body).toEqual({
      title: 'Jane Doe',
      email: 'jane@raystack.org',
      avatar: 'data:image/png;base64,new',
      name: 'jane_doe'
    });
    expect(body).not.toHaveProperty('metadata');
  });

  it('leaves out metadata when the title changes', () => {
    const formValues = { ...storedUser, title: 'Jane Smith' };

    const body = profileUpdateBody(formValues, storedUser.name);

    expect(body.title).toBe('Jane Smith');
    expect(body).not.toHaveProperty('metadata');
  });
});
