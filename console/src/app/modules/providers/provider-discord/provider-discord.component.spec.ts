import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';

import { ProviderDiscordComponent } from './provider-discord.component';

describe('ProviderDiscordComponent', () => {
  let component: ProviderDiscordComponent;
  let fixture: ComponentFixture<ProviderDiscordComponent>;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      declarations: [ProviderDiscordComponent],
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ProviderDiscordComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});