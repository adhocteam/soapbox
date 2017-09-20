FROM ruby:2.4.1
RUN apt-get update -qq && apt-get install -y build-essential nodejs
RUN mkdir /web
COPY . /web
WORKDIR /web
RUN bundle install
CMD ["bundle", "exec", "rails", "s", "-b", "0.0.0.0"]
